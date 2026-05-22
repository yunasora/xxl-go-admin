package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type RegisterCenter struct {
	store sync.Map
	//记录轮循次数
	routeCount sync.Map
}

func (rc *RegisterCenter) StoreNode(appName string, addr string) {
	actual, _ := rc.store.LoadOrStore(appName, &sync.Map{})
	nodeMap := actual.(*sync.Map)
	nodeMap.Store(addr, time.Now())
}

// 全局实例
var RegsC = &RegisterCenter{}

func (rc *RegisterCenter) ElectNode(appName string) (string, error) {

	actual, ok := rc.store.Load(appName)
	if !ok {
		return "", fmt.Errorf("没有找到任何属于 [%s] 的任何执行器", appName)
	}
	nodeMap := actual.(*sync.Map)

	//查找还活着的机器
	var onlinenodes []string
	nodeMap.Range(func(key, value interface{}) bool {
		addr := key.(string)
		lastHeartbeat := value.(time.Time)

		if time.Since(lastHeartbeat) > 90*time.Second {
			return true
		}

		onlinenodes = append(onlinenodes, addr)
		return true
	})

	if len(onlinenodes) == 0 {
		return "", fmt.Errorf("属于 [%s] 的执行器全都超时下线了", appName)
	}

	//开始轮循
	var currentCount int64
	if count, loaded := rc.routeCount.LoadOrStore(appName, int64(0)); loaded {
		currentCount = count.(int64)
	}
	nextCount := currentCount + 1
	rc.routeCount.Store(appName, nextCount)

	//轮循%机器数，决定用谁
	index := int(currentCount % int64(len(onlinenodes)))
	chosenAddr := onlinenodes[index]

	log.Printf("[负载均衡][%s], || 当前机器在线数: %d ||, 使用第[%d]个机器 : %s", appName, len(onlinenodes), index+1, chosenAddr)
	return chosenAddr, nil
}

// 清理死亡下线机器
func (rc *RegisterCenter) StartClearloop() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			rc.store.Range(func(appNameKey, appNameVal interface{}) bool {
				appName := appNameKey.(string)
				nodeMap := appNameVal.(*sync.Map)

				nodeMap.Range(func(addrKey, addrVal interface{}) bool {
					addr := addrKey.(string)
					lastHeartbeat := addrVal.(time.Time)

					if time.Since(lastHeartbeat) > 90*time.Second {
						log.Printf("[死亡清理] 执行器[%s] 节点:[%s] 长期未发送消息,予以剔除", appName, addr)
						nodeMap.Delete(addrKey)
					}
					return true
				})
				return true
			})
		}
	}()
}
