import json
import requests
import sys
import os
from difflib import Differ
from pprint import pprint

queryfile = sys.argv[1]
endpoints = sys.argv[2:]
apikey = os.environ.get('TRANSITLAND_API_KEY')

reqs = []
with open(queryfile, encoding='utf-8') as f:
    for row in f.readlines():
        reqs.append(json.loads(row))

for count,req in enumerate(reqs):
    print("q:", count)
    b = req['body']
    resps = []
    for ep in endpoints:
        print("\t", ep)
        resp = requests.post(ep, json = b, headers={'apikey':apikey}).json()
        print("\t\tok")
        resps.append(resp)

    for i in range(1,len(resps)):
        r1 = resps[i]
        r2 = resps[i-1]
        if r1 != r2:
            print("diff:")
            d = Differ()
            text1 = json.dumps(r1, indent=2).splitlines()
            text2 = json.dumps(r2, indent=2).splitlines()
            result = list(d.compare(text1, text2))
            sys.stdout.writelines(result)
                