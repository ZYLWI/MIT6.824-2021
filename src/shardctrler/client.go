package shardctrler

//
// Shardctrler clerk.
//

import "6.824/labrpc"
import "crypto/rand"
import "math/big"

type Clerk struct {
	servers   []*labrpc.ClientEnd
	leaderId  int64
	clientId  int64 // generated by nrand(), it would be better to use some distributed ID generation algorithm that guarantees no conflicts
	commandId int64 // (clientId, commandId) defines a operation uniquely
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	return bigx.Int64()
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	return &Clerk{
		servers:   servers,
		leaderId:  0,
		clientId:  nrand(),
		commandId: 0,
	}
}

func (ck *Clerk) Query(num int) Config {
	return ck.Command(&CommandRequest{Num: num, Op: OpQuery})
}

func (ck *Clerk) Join(servers map[int][]string) {
	ck.Command(&CommandRequest{Servers: servers, Op: OpJoin})
}

func (ck *Clerk) Leave(gids []int) {
	ck.Command(&CommandRequest{GIDs: gids, Op: OpLeave})
}

func (ck *Clerk) Move(shard int, gid int) {
	ck.Command(&CommandRequest{Shard: shard, GID: gid, Op: OpMove})
}

//
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("ShardCtrler.Command", &request, &response)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Command(request *CommandRequest) Config {
	request.ClientId, request.CommandId = ck.clientId, ck.commandId
	for {
		var response CommandResponse
		if !ck.servers[ck.leaderId].Call("ShardCtrler.Command", request, &response) || response.Err == ErrWrongLeader || response.Err == ErrTimeout {
			ck.leaderId = (ck.leaderId + 1) % int64(len(ck.servers))
			continue
		}
		ck.commandId++
		return response.Config
	}
}
