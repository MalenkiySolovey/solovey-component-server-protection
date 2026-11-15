import fs from 'node:fs'
import path from 'node:path'
import assert from 'node:assert/strict'
import {createHash} from 'node:crypto'
import {execFileSync} from 'node:child_process'
import {fileURLToPath} from 'node:url'

const root=fileURLToPath(new URL('../',import.meta.url))
const hash=bytes=>createHash('sha256').update(bytes).digest('hex')
const git=args=>execFileSync('git',args,{cwd:root,encoding:'utf8',maxBuffer:32*1024*1024}).trim()
assert.equal(process.env.GOWORK,'off','qualification requires GOWORK=off')
assert.ok((process.env.GOFLAGS??'').split(/\s+/).every(flag=>!flag||flag==='-mod=readonly'),'source qualification rejects Go flag substitution')
const identity='github.com/MalenkiySolovey/solovey-component-server-protection'
const mod=JSON.parse(execFileSync('go',['mod','edit','-json'],{cwd:root,encoding:'utf8'}))
assert.equal(mod.Module.Path,identity)
assert.ok(!(mod.Replace??[]).length,'source module replacements are forbidden')
assert.deepEqual(Object.fromEntries((mod.Require??[]).map(item=>[item.Path,item.Version])),{'github.com/MalenkiySolovey/solovey-host':'v0.9.0','golang.org/x/sys':'v0.49.0'},'unreviewed module requirements')
const source=JSON.parse(fs.readFileSync(path.join(root,'SOURCE_PROVENANCE.json')))
assert.equal(source.schema,'solovey.component.source-provenance/v1')
assert.equal(source.component,'server-protection')
assert.equal(source.originalCommit,'1b6c5818c5bb0f03c879ec5580f967a5aa188396')
assert.equal(hash(fs.readFileSync(path.join(root,'LICENSE'))),source.licenseSHA256,'original license changed')
git(['merge-base','--is-ancestor',source.filteredHistoryCommit,'HEAD'])
const checked=new Set(['SOURCE_PROVENANCE.json'])
function verify(name,expected){
 assert.ok(name&&!name.includes('..')&&!path.isAbsolute(name))
 assert.equal(hash(fs.readFileSync(path.join(root,name))),expected,`source drift: ${name}`)
 assert.ok(!checked.has(name),`duplicate source mapping: ${name}`)
 checked.add(name)
}
for(const item of source.files)verify(item.destination,item.destinationSHA256)
for(const [name,expected] of Object.entries(source.authoredFiles))verify(name,expected)
const tracked=git(['ls-files','-z']).split('\0').filter(Boolean)
assert.deepEqual([...checked].sort(),[...tracked].sort(),'every tracked file needs a reviewed source identity')
assert.ok(!tracked.some(name=>name.includes('node_modules/')||name.startsWith('.frontend-check/')||name.startsWith('components/')),'generated or retired Product source included')
const output=execFileSync('go',['list','-deps','-test','-json','./...'],{cwd:root,encoding:'utf8',maxBuffer:24*1024*1024})
const packages=(output.match(/^{[\s\S]*?^}/gm)??[]).map(JSON.parse)
assert.ok(packages.length>0)
const allowed=new Set([identity,'github.com/MalenkiySolovey/solovey-host','golang.org/x/sys'])
for(const pkg of packages)if(pkg.Module){assert.ok(allowed.has(pkg.Module.Path),`unapproved dependency ${pkg.ImportPath}`);assert.ok(!pkg.Module.Replace,'source module substitution')}
for(const name of tracked.filter(name=>/\.(go|ts|vue)$/.test(name))) {
 const text=fs.readFileSync(path.join(root,name),'utf8')
 assert.ok(!text.includes('github.com/MalenkiySolovey/solovey-ui/'),`private Product dependency: ${name}`)
 assert.ok(!/from\s+['"]@\//.test(text),`private UI import: ${name}`)
 assert.ok(!/\b(?:M[026][A-C]?|R7|Wave|Phase|Goal)[-_]\w+/.test(text),`development-stage identifier: ${name}`)
}
const descriptor=JSON.parse(fs.readFileSync(path.join(root,'component.json')))
assert.equal(descriptor.id,'server-protection')
assert.equal(descriptor.delivery,'in-process')
assert.equal(descriptor.defaultEnabled,false)
assert.equal(descriptor.version,'2')
assert.equal(descriptor.database.migrationVersion,'2')
assert.equal(descriptor.database.tables.length,30)
assert.deepEqual(descriptor.database.compatibilityRange,{minimumSchema:'1',maximumSchema:'1'})
assert.equal(hash(fs.readFileSync(path.join(root,'component.json'))),source.manifestSHA256)
// Review the complete retained history and current source; report object IDs only.
const objects=git(['rev-list','--objects','HEAD']).split('\n').filter(Boolean).map(line=>line.split(' ')[0])
const batch=execFileSync('git',['cat-file','--batch'],{cwd:root,input:objects.join('\n')+'\n',maxBuffer:128*1024*1024})
const rules=[/-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----/,/\b(?:gh[pousr]_[A-Za-z0-9]{36,255}|github_pat_[A-Za-z0-9_]{50,255})\b/,/\b(?:AKIA|ASIA)[A-Z0-9]{16}\b/,/\bxox[baprs]-[A-Za-z0-9-]{24,}\b/]
const findings=[];let offset=0,blobs=0
while(offset<batch.length){
 const end=batch.indexOf(10,offset);assert.ok(end>offset)
 const [oid,type,sizeText]=batch.subarray(offset,end).toString().split(' '),size=Number(sizeText)
 assert.ok(Number.isSafeInteger(size)&&size>=0)
 const bytes=batch.subarray(end+1,end+1+size);offset=end+size+2
 if(type==='blob'||type==='commit'){if(type==='blob')blobs++;const text=bytes.toString();if(rules.some(rule=>rule.test(text)))findings.push({object:oid,type})}
}
assert.deepEqual(findings,[],'potential secret in retained source history')
for(const name of tracked){const text=fs.readFileSync(path.join(root,name),'utf8');assert.ok(!rules.some(rule=>rule.test(text)),`potential secret in current source: ${name}`)}
console.log(JSON.stringify({status:'PASS',packages:packages.length,allowedModules:[...allowed],mappedSources:source.files.length,trackedFiles:tracked.length,privateProductImports:0,historyBlobsReviewed:blobs,secretFindings:0,filteredHistoryPreserved:true}))
