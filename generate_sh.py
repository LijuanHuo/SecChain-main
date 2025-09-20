import os,sys
os.chdir(sys.path[0])

NodeNum = 4
ShardNum = 32
m = 3

with open(f'bat_sn={ShardNum}_nn={NodeNum}.sh', 'w') as f:
    for i in range(0, ShardNum):
        for j in range(1, NodeNum):
            f.write(f"go run main.go -n {j} -N {NodeNum} -s {i} -S {ShardNum} -m {m} &\n\n")
    
    f.write(f"go run main.go -c -N {NodeNum} -S {ShardNum} -m {m} &\n\n")

    for k in range(0, ShardNum):
        f.write(f"go run main.go -n 0 -N {NodeNum} -s {k} -S {ShardNum} -m {m} &\n\n")
