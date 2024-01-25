import json
import requests
import sys
import os
import time
from difflib import Differ, context_diff
from pprint import pprint

tnow = time.time()
queryfile = sys.argv[1]
endpoints = sys.argv[2:]
apikey = os.environ.get('TRANSITLAND_API_KEY')

reqs = []
with open(queryfile, encoding='utf-8') as f:
    for row in f.readlines():
        reqs.append(json.loads(row))

for count,req in enumerate(reqs):
    print("=============================")
    print("q:", count)
    b = req.get('body')
    if not b:
        continue
    print(b.get('query'))
    print(b.get('variables'))
    resps = []
    for ep in endpoints:
        print("\t", ep)
        t = time.time()
        resp = requests.post(ep, json = b, headers={'apikey':apikey}).json()
        print("\t\tok", time.time()-t)
        resps.append(resp)


    ok = True
    for i in range(1,len(resps)):
        r1 = resps[i]
        r2 = resps[i-1]
        if r1 != r2:
            print("diff:")
            text1 = json.dumps(r1, indent=2).splitlines()
            text2 = json.dumps(r2, indent=2).splitlines()
            result = list(context_diff(text1, text2))
            # pprint(result)
            print("\n".join(result))
            ok = False    

    if not ok:
        for i,resp in enumerate(resps):
            with open(f"q-result-{tnow}-{count}-{i}.json", "w", encoding="utf-8") as outf:
                json.dump(resp, outf, indent=2)
            