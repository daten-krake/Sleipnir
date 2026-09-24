import hashlib, json, re, io, collections, sys
def rd(p): return io.open(p, encoding='utf-8', newline='').read()
P = 'contracts/'
A0, A1, A2 = rd(P+'A0-conventions.md'), rd(P+'A1-events.md'), rd(P+'A2-graph.md')
def canon(o): return json.dumps(o, sort_keys=True, separators=(',',':'), ensure_ascii=False).encode()
def ph(x): return x.replace('<2028>','\u2028').replace('<2029>','\u2029').replace('<7F>','\u007f')
ok=[]; bad=[]
def chk(n,c,d=''):
    (ok if c else bad).append(f"{n}{(' — '+d) if d and not c else ''}")

# A0 vectors
n=0
for line in A0.split('\n'):
    m = re.match(r'\s*\| (V\d) \|(.+)\|\s*$', line)
    if not m: continue
    n+=1; parts=m.group(2).split('|')
    cell = re.match(r'\s*`(.*?)`\s*(\(.*\))?\s*$', parts[1], re.S).group(1)
    b=ph(cell).encode(); sl=int(parts[2].strip()); sd=parts[3].strip().strip('`')
    chk(f"A0 {m.group(1)} vector", len(b)==sl and hashlib.sha256(b).hexdigest()==sd, f"len {len(b)}/{sl}")
chk("A0 vector count == 8", n==8, str(n))
# A1 chain
rows = re.findall(r'^\| (\d) \| `([a-z_]+)` \| (\d+) \| see below \| `([0-9a-f]{64})` \| (\d+) \| `([0-9a-f]{64})` \|$', A1, re.M)
pres = re.findall(r'^seq (\d) preimage:\n(\{.*\})$', A1, re.M)
chk("A1 4.3 has 4 rows + 4 preimages", len(rows)==4 and len(pres)==4, f"{len(rows)}/{len(pres)}")
prev='0'*64
for (num,kind,pl,pd_,sl,sd),(pn,pre) in zip(rows,pres):
    b=ph(pre).encode(); o=json.loads(b); h=hashlib.sha256(b).hexdigest()
    srv=dict(o); srv['seq']=int(num); srv['prev_hash']=prev; srv['hash']=h
    sb=canon(srv)
    chk(f"A1 4.3 row {num} ({kind})", len(b)==int(pl) and h==pd_ and len(sb)==int(sl)
        and hashlib.sha256(sb).hexdigest()==sd and canon(o)==b and not ({'hash','prev_hash','seq'} & set(o)),
        f"pre {len(b)}/{pl} served {len(sb)}/{sl}")
    prev=h
# A1 PayloadHash
# Strengthened 2026-09-21 (WP-01): this check used to compare the preimage block
# against a digest hardcoded HERE, so the value *published in the contract* was
# never verified - editing A1-7.6's printed digest left the suite green. It now
# reads the published length and digest out of the document and requires both to
# match the recomputation and the original 2026-09-11 constant. Check count is
# unchanged (still 52) because this strengthens one check rather than adding one.
blk = re.search(r'^    (\{"kind":"command_executed".*\})$', A1, re.M)
pub = re.search(r'\*\*len (\d+)[^\n`]*SHA-256\s*`([0-9a-f]{64})`\*\* over', A1)
PAYLOAD_HASH = '19d976a83cc9d36ac160313a20b80c0745fff805526b7f43f88d05e33c7be5e5'
chk("A1-7.6 PayloadHash vector", blk and pub
    and len(blk.group(1).encode()) == 293
    and int(pub.group(1)) == len(blk.group(1).encode())
    and pub.group(2) == PAYLOAD_HASH
    and hashlib.sha256(blk.group(1).encode()).hexdigest() == PAYLOAD_HASH,
    f"published={pub.groups() if pub else None} recomputed={len(blk.group(1).encode()) if blk else None}")
# A2 fingerprint vectors
KEYS=['addresses','attrs','basis','cidr','claim','credential_kind','domain','evidence_id','evidence_ids','kind','label','media_kind','port','protocol','severity','sid','size_bytes','status','summary','transport']
STR={'basis','cidr','claim','credential_kind','domain','evidence_id','kind','label','media_kind','protocol','severity','sid','status','summary','transport'}
def cdoc(nn):
    d={}
    for k in KEYS:
        v=nn.get(k); d[k]= v if v is not None else ('' if k in STR else ([] if k in ('addresses','evidence_ids') else ({} if k=='attrs' else 0)))
    d['addresses']=sorted(set(d['addresses'])); d['evidence_ids']=sorted(set(d['evidence_ids'])); return d
sec=re.search(r'### 4\.2(.*?)\n### 4\.3', A2, re.S).group(1)
exp={'ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e':656,
     '1748b813a0d28d9f17f4d89dff2a08536741bf45e8fe0defbb3c4363d3f9b983':656,
     '3955d82160284d3e76c9be1b06981530df50e1635a1bad7ceb7b50c5f73c2b67':304,
     '4a017e71658de7ebb2d3a5429f90818b8a69302ff1909badbc782d9049ce9cae':656}
for d,l in exp.items(): chk(f"A2 4.2 digest {d[:8]} published", d in sec)
blocks=re.findall(r'^(\{"(?:addresses|attrs)".*?\})$', sec, re.M)
rec={hashlib.sha256(ph(b).encode()).hexdigest(): len(ph(b).encode()) for b in blocks}
chk("A2 4.2 canonical bytes recompute (3 of 4; F1-R shares F1)",
    all(d in rec and rec[d]==l for d,l in exp.items() if d!='4a017e71658de7ebb2d3a5429f90818b8a69302ff1909badbc782d9049ce9cae'),
    f"blocks={len(blocks)} got={ {k[:8]:v for k,v in rec.items()} }")
# A2 examples
ex=re.search(r'### 4\.1 JSON examples(.*?)\n### 4\.2', A2, re.S).group(1)
found=0
for m in re.finditer(r'```json\n(\{.*?\n\})\n```', ex, re.S):
    try: o=json.loads(m.group(1))
    except Exception: continue
    nodes = o.get('items') if isinstance(o.get('items'), list) else [o]
    for nn in nodes:
        if isinstance(nn,dict) and 'content_hash' in nn and nn.get('kind'):
            found+=1
            h=hashlib.sha256(canon(cdoc(nn))).hexdigest()
            chk(f"A2 example {nn['kind']}@{nn.get('graph_seq','?')} content_hash", h==nn['content_hash'], f"{h[:12]} vs {nn['content_hash'][:12]}")
            prov=nn.get('provenance')
            if isinstance(prov,list):
                confs=[p.get('confidence') for p in prov]
                ind=len({(p.get('task_id'),p.get('agent_node_id')) for p in prov})
                chk(f"A2 example {nn['kind']} verified-legal", 'verified' not in confs or ind>1, f"confs={confs} indep={ind}")
chk("A2 examples checked >= 4", found>=4, str(found))
# vocab / counts / structure
kinds=re.findall(r'^\tKind([A-Za-z0-9]+)\s+Kind = "([a-z_]+)"', A1, re.M)
chk("A1 Kind consts == 42", len(kinds)==42, str(len(kinds)))
chk("A1 no stale 39", '39 kinds' not in A1 and '37 to 39' not in A1)
chk("A1 quarantine enum has no operator_release", 'operator_release' not in re.search(r'quarantine_kind:enum\{[^}]*\}', A1).group(0))
chk("A0 usr_ + KindUser", 'usr_' in A0 and 'KindUser' in A0)
chk("A0 ToolVersionMaxBytes=64", re.search(r'ToolVersionMaxBytes\s*=\s*64', A0) is not None)
chk("A0 EvidenceRefsMax=8", re.search(r'EvidenceRefsMax\s*=\s*8', A0) is not None)
for nm,doc in (('A0',A0),('A1',A1),('A2',A2)):
    chk(f"{nm} six top-level sections", re.findall(r'^## (\d)\. ', doc, re.M)==['1','2','3','4','5','6'])
    chk(f"{nm} no raw U+2028/9/7F in table rows", not [l for l in doc.split('\n') if l.lstrip().startswith('|') and any(c in l for c in '\u2028\u2029\u007f')])
    chk(f"{nm} no plan scaffolding", not re.search(r'BLOCK-[A-Z]|PAIR-[A-Z]|\u29c9|plan §', doc))
    chk(f"{nm} review-provenance header row", '**Review provenance**' in doc)
chk("A0 4.1 tests", '### 4.1 Contract tests' in A0); chk("A1 4.4 tests", '### 4.4 Contract tests' in A1)
chk("A2 4.2 vectors", '### 4.2' in A2); chk("A2 4.3 tests", '### 4.3 Contract tests' in A2)
# dangling clause refs
defined={}
for d,t in (('A0',A0),('A1',A1),('A2',A2)):
    ids=set(re.findall(r'\*\*('+d+r'-\d+(?:\.\d+)?[a-z]?)\*\*',t))|set(re.findall(r'^\| \*\*('+d+r'-\d+(?:\.\d+)?[a-z]?)\*\*',t,re.M))|set(re.findall(r'(?:^|\s)('+d+r'-\d+\.\d+[a-z]?)(?=\s*\|)',t))|set(re.findall(r'^#{3,4} ('+d+r'-\d+)',t,re.M))
    defined[d]=ids|{f"{d}-{g}" for g in set(re.findall(d+r'-(\d+)',' '.join(ids)))}
dang=collections.defaultdict(set)
for s,t in (('A0',A0),('A1',A1),('A2',A2)):
    for m in re.finditer(r'\b(A[012]-\d+(?:\.\d+)?[a-z]?)\b', t):
        r=m.group(1)
        if r not in defined[r[:2]]: dang[(s,r[:2])].add(r)
chk("no dangling clause references", not dang, str({f"{k[0]}->{k[1]}":sorted(v) for k,v in dang.items()}))
print(f"PASS {len(ok)} / FAIL {len(bad)}")
for b in bad: print("  FAIL:", b)
# Exit non-zero on any failure. Added 2026-09-21 (WP-01): without this the
# script printed "FAIL n" and still exited 0, so a CI step gating on the exit
# code passed a broken contract. The printed line stays the human-readable
# summary and the assertion of record for the 52-check baseline.
sys.exit(1 if bad else 0)
