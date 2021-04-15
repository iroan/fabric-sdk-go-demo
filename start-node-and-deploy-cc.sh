ccPath=${PWD}/chaincode
cd /mnt/wd40ezrz/gitrepo/interested/fabric-samples/test-network

source ./scripts/envVar.sh
source ./scripts/utils.sh

export COMPOSE_PROJECT_NAME="fhe"
export CCNAME="demo02"

./network.sh down -i 2.2
./network.sh up createChannel -i 2.2


export PATH=$PATH:/mnt/wd40ezrz/gitrepo/interested/fabric/build/bin
export FABRIC_CFG_PATH=/mnt/wd40ezrz/gitrepo/interested/fabric-samples/config
export VERBOSE=true

./network.sh deployCC -ccn ${CCNAME} -ccp ${ccPath} -ccl go