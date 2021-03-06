# 分布式缓存中一致性哈希算法的应用

这里就不细说什么是分布式缓存，主要写下什么是一致性哈希算法以及它的原理，为什么使用一致性哈希算法。 先来一道场景题，如果有n个缓存节点，如何保证key可以均匀分布在各个节点，以及下次可以通过key来命中该缓存？
大家一开始肯定想到的是`hash`算法了，就来聊聊`hash`

## 哈希取模

```
    inedx = hash(key)%n 
```

这就是哈希取模，使用这种方式确实可以在多个缓存(主机)之间进行数据分布以及查找。但也会存在一些问题。

* 可拓展性较差
* 容错率低

现假设有一台R服务器宕机了，那么为了填补空缺，要将宕机的服务器从编号列表中移除，后面的服务器按顺序前移一位并将其编号值减一，此时每个key就要按 `index = hash(key) % (n-1)`重新计算。
同样，如果新增一台服务器，规则也同样需要重新计算，`index = hash(key) % (n+1)`。因此，系统中如果有服务器更变，会直接影响到Hash值，
大量的key会重定向到其他服务器中，造成缓存命中率降低，而这种情况在分布式系统中是十分糟糕的。在请求量较大的情况下，就会造成缓存雪崩。

## 一致性哈希

### 算法原理

一致性哈希算法将所有生成的`hashcode`构建成一个环。范围在0～2^32-1。

![](https://i.loli.net/2019/05/08/5cd1b9fe31560.jpg)

* 可以使用服务器IP+端口或者编号进行哈希，放置到环上。
* 计算key的hashcode，放置在环上，顺时针寻找到的第一个节点，就是应选取的节点/机器。

如下图，将各个节点、主机的唯一ID（host，port）进行hash，放置到环上得到，N1、N2、N3。 之后需要将数据定位到对应的节点上，使用同样的hash函数将Key也映射到这个环上。 这样按照顺时针方向就可以把 k1 定位到
N1节点，k2 定位到 N3节点，k3 定位到 N2节点 。

![](https://i.loli.net/2019/05/08/5cd1ba05955b9.jpg)

### 增删节点的问题解决

新增一个节点如下，在N3与N2节点之间，那么下一次查找K3缓存的时候，由于K4主机的存在，顺时针hash到K4节点，只有K3受到影响，其他数据不会受到影响。
在实际的场景下，只会影响到节点附近小部分的缓存。删除节点也是类似的，举一反三，思考一下就明白了。

![](https://i.loli.net/2019/05/08/5cd1ba0c7519c.jpg)

### 数据倾斜问题解决

如果服务器的节点过少，容易引起 key 的倾斜。也就是key在各个节点之间负载不均衡。

为了解决这个问题，引入了虚拟节点的概念，一个真实节点对应多个虚拟节点。虚拟节点扩充了节点的数量，解决了节点较少的情况下数据容易倾斜的问题。而且代价非常小，只需要增加一个字典(map)维护真实节点与虚拟节点的映射关系即可。


