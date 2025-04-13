mkdir bully-leader-electaion 

cd bully-leader-electaion 

go mod init codecraftwithreza/ble

go get -u github.com/gin-gonic/gin



endpoints: 
    1. ping
    2. electation
    3. leader-elected


main functions:
    1. comunicateWithPeer
    2. elect
    3. pingContinuslyLeader
