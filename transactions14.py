import csv
import os,sys
os.chdir(sys.path[0])

import json

SplitNum = 1000

rows = []

with open('14000000to14249999_ERC20Transaction.csv', 'r') as file:
    reader = csv.reader(file)
    i = 0
    for row in reader:
        rows.append(row)
        i += 1
        if(i==SplitNum+1):
            break

save_name = f"TestTx_{SplitNum}.csv"

# {
#     "index": "TX624412",
#     "md5": "————",
#     "userName": "a2f87b145bc6b1c0844c5679d4e1a0d9adbe3e32",
#     "recipientName": "e7d8e926f8a76563aebd1cfecbda3bdf04eab528",
#     "amount": 200,
#     "status": "success",
#     "isVerified": false,
#     "date": "2024-11-07"
# }

save_name = "transactions.json"

with open(save_name, 'w', encoding='utf-8') as sf:
    # writer = csv.writer(sf)
    ls = []
    i = 0
    for row in rows:
        i += 1
        if(i==1):
            continue
        dir0 = {}
        dir0["index"] = f"TX{i:06}"
        dir0["md5"] =  "————"
        dir0["userName"] = row[4][2:]
        dir0["recipientName"] = row[5][2:]
        strtmp = row[8]
        if(len(row[8]) > 6):
            strtmp = row[8][0:6]
        
        print(strtmp)
        dir0["amount"] = int(strtmp) % 997
        dir0["status"] =  "success"
        dir0["isVerified"] =  False
        date = dir0["amount"] % 16
        dir0["date"] = "2024-11-" + f"{date:02}"
        ls.append(dir0)
    json.dump(ls, sf, ensure_ascii=False, indent=4)



        