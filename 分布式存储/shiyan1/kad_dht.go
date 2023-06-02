package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"sort"
	"strconv"
)

// 最大K桶大小
const K = 3

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

// 打印桶中的内容
func (b *Bucket) PrintContents() {
	for _, node := range b.Nodes {
		fmt.Println("nodeID = ",node.ID)
	}
}

type Peer struct {
	ID     *big.Int
	Bucket *Bucket
}

func NewPeer(id string) *Peer {
	h := sha1.New()
	h.Write([]byte(id))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)
	return &Peer{ID: intID, Bucket: NewBucket()}
}

// 将节点加入到Peer的KBucket中
func (p *Peer) InsertNode(nodeID string) {
	h := sha1.New()
	h.Write([]byte(nodeID))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)
	p.Bucket.Insert(Node{ID: intID})
}

// 打印Peer的KBucket的内容
func (p *Peer) PrintBucketContents() {
	p.Bucket.PrintContents()
}

// 寻找节点，如果节点存在返回true，否则返回false
func (p *Peer) FindNode(nodeID string) bool {
	h := sha1.New()
	h.Write([]byte(nodeID))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)
	for _, node := range p.Bucket.Nodes {
		if node.ID.Cmp(intID) == 0 {
			return true
		}
	}
	return false
}

type Network struct {
	Peers map[string]*Peer
}

func NewNetwork() *Network {
	return &Network{Peers: make(map[string]*Peer)}
}

// 添加Peer到网络中
func (n *Network) AddPeer(p *Peer) {
	n.Peers[p.ID.String()] = p
}

// 找到最近的两个Peer
func (n *Network) FindClosestPeers(nodeID string) []*Peer {
	h := sha1.New()
	h.Write([]byte(nodeID))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)
	distances := make([]*big.Int, 0, len(n.Peers))
	for _, peer := range n.Peers {
		distances = append(distances, distance(peer.ID, intID))
	}
	sort.Slice(distances, func(i, j int) bool { return distances[i].Cmp(distances[j]) < 0 })
	closestPeers := make([]*Peer, 0, 2)
	for _, dist := range distances[:2] {
		for _, peer := range n.Peers {
			if distance(peer.ID, intID).Cmp(dist) == 0 {
				closestPeers = append(closestPeers, peer)
			}
		}
	}
	return closestPeers
}

// 广播节点信息
func (n *Network) BroadcastNode(nodeID string) {
	h := sha1.New()
	h.Write([]byte(nodeID))
	bs := h.Sum(nil)
	intID := new(big.Int)
	intID.SetBytes(bs)
	for _, peer := range n.Peers {
		peer.InsertNode(intID.String())
	}
}

// TestInsertNode 测试插入功能
func TestInsertNode() {
	peer := NewPeer("test")
	for i := 0; i < 5; i++ {
		peer.InsertNode(strconv.Itoa(i))
	}
	fmt.Println("测试插入成功")
}

// TestPrintBucketContents 测试打印功能
func TestPrintBucketContents() {
	peer := NewPeer("test")
	for i := 0; i < K; i++ {
		peer.InsertNode(strconv.Itoa(i))
	}
	fmt.Println("test ID:", peer.ID)
	peer.PrintBucketContents()
	fmt.Println("测试打印成功")
	// This is a visual test. Ensure the printout is correct.
}
func main() {
	net := NewNetwork()

	// 初始化5个Peer
	for i := 0; i < 5; i++ {
		p := NewPeer(fmt.Sprintf("%x", i))
		net.AddPeer(p)
	}

	// 加入200个新的Peer
	for i := 5; i < 205; i++ {
		p := NewPeer(fmt.Sprintf("%x", i))
		net.AddPeer(p)
		net.BroadcastNode(p.ID.String())
	}

	// 打印所有Peer的KBucket的内容
	for _, peer := range net.Peers {
		fmt.Println("Peer ID:", peer.ID)
		peer.PrintBucketContents()
		fmt.Println()
	}

	//测试
	TestInsertNode()
	TestPrintBucketContents()
}
