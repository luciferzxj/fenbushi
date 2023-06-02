package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

// Bucket size
const K = 3

// Maximum number of Buckets
const N = 160

type Node struct {
	ID *big.Int
}

type Bucket struct {
	Nodes []Node
}

func NewBucket() *Bucket {
	return &Bucket{Nodes: make([]Node, 0)}
}

// 使用big.Int实现ID之间的距离计算
func distance(a, b *big.Int) *big.Int {
	return new(big.Int).Xor(a, b)
}

// 插入节点到桶中
func (b *Bucket) Insert(n Node) {
	for _, node := range b.Nodes {
		if node.ID.Cmp(n.ID) == 0 {
			return
		}
	}
	if len(b.Nodes) < K {
		b.Nodes = append(b.Nodes, n)
	} else {
		// 这里简单地删除最旧的节点
		b.Nodes = b.Nodes[1:]
		b.Nodes = append(b.Nodes, n)
	}
}

type DHT struct {
	Buckets [N]*Bucket
}

type Peer struct {
	ID   *big.Int
	DHT  *DHT
	Data map[string][]byte
}

func NewPeer(id string) *Peer {
	h := sha1.New()
	h.Write([]byte(id))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)

	dht := &DHT{}
	for i := range dht.Buckets {
		dht.Buckets[i] = NewBucket()
	}

	return &Peer{ID: intID, DHT: dht, Data: make(map[string][]byte)}
}

type Network struct {
	Peers map[string]*Peer
}

func NewNetwork() *Network {
	return &Network{Peers: make(map[string]*Peer)}
}

func (n *Network) AddPeer(p *Peer) {
	n.Peers[p.ID.String()] = p
}

func (n *Network) FindClosestPeers(id *big.Int) []*Peer {
	distances := make([]*big.Int, 0, len(n.Peers))
	for _, peer := range n.Peers {
		distances = append(distances, distance(peer.ID, id))
	}
	sort.Slice(distances, func(i, j int) bool { return distances[i].Cmp(distances[j]) < 0 })
	closestPeers := make([]*Peer, 0, 2)
	for _, dist := range distances[:2] {
		for _, peer := range n.Peers {
			if distance(peer.ID, id).Cmp(dist) == 0 {
				closestPeers = append(closestPeers, peer)
			}
		}
	}
	return closestPeers
}

func (p *Peer) SetValue(key, value []byte, network *Network) bool {
	h := sha1.New()
	h.Write(value)
	hashedValue := h.Sum(nil)
	if string(key) != string(hashedValue) {
		return false
	}
	if _, ok := p.Data[string(key)]; ok {
		return true
	}

	// 算出这个节点对应的桶
	intKey := new(big.Int)
	intKey.SetBytes(key)
	dist := distance(p.ID, intKey)
	index := dist.BitLen() - 1
	if index < 0 {
		index = 0
	}

	// 从对应的桶里面选择2个距离Key最近的节点
	closestPeers := network.FindClosestPeers(intKey)

	// 对这2个节点执行SetValue操作
	for _, peer := range closestPeers {
		peer.Data[string(key)] = value
	}

	p.Data[string(key)] = value
	return true
}

func (p *Peer) GetValue(key []byte, network *Network) []byte {
	if value, ok := p.Data[string(key)]; ok {
		return value
	}

	// 如果自己没有存储当前Key
	// 对当前的Key执行一次FindNode操作，找到距离当前Key最近的2个Peer
	intKey := new(big.Int)
	intKey.SetBytes(key)
	closestPeers := network.FindClosestPeers(intKey)

	// 对这两个Peer执行GetValue操作
	for _, peer := range closestPeers {
		if value, ok := peer.Data[string(key)]; ok {
			// 一旦有一个节点返回value，则返回校验成功之后的value
			h := sha1.New()
			h.Write(value)
			hashedValue := h.Sum(nil)
			if string(key) == string(hashedValue) {
				return value
			}
		}
	}

	// 否则返回nil
	return nil
}
func generateRandomString() string {
	b := make([]byte, 50)
	if _, err := rand.Read(b); err != nil {
		println(err)
	}
	return fmt.Sprintf("%x", b)
}

func getRandomPeer(net *Network) *Peer {
	peerIDs := make([]string, 0, len(net.Peers))
	for id := range net.Peers {
		peerIDs = append(peerIDs, id)
	}
	randomPeerIndex := rand.Intn(len(peerIDs)) // 随机选一个Peer
	return net.Peers[peerIDs[randomPeerIndex]]
}
func main() {
	rand.Seed(time.Now().UnixNano())

	net := NewNetwork()

	// 初始化100个Peer
	for i := 0; i < 100; i++ {
		p := NewPeer(fmt.Sprintf("%x", i))
		net.AddPeer(p)
	}

	// 随机生成200个字符串
	strings := make([]string, 200)
	for i := range strings {
		strings[i] = generateRandomString()
	}

	// 随机从100个节点中选出一个执行SetValue(Key,Value)操作
	keys := make([][]byte, 200)
	for i, str := range strings {
		hash := sha1.New()
		hash.Write([]byte(str))
		bs := hash.Sum(nil)
		key := bs[:20] // 取前20个字节作为Key
		keys[i] = key
		value := []byte(str)

		peer := getRandomPeer(net)

		peer.SetValue(key, value, net)
	}

	// 从200个Key中随机选择100个，然后每个Key再去随机找一个节点调用GetValue操作
	for i := 0; i < 100; i++ {
		randomKeyIndex := rand.Intn(200) // 随机选一个Key
		key := keys[randomKeyIndex]

		peer := getRandomPeer(net)

		value := peer.GetValue(key, net)
		if value != nil {
			fmt.Printf("GetValue(%x) = %s\n", key, value)
		} else {
			fmt.Printf("GetValue(%x) = nil\n", key)
		}
	}
}


