package vyos

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
)

const (
	ENDPOINT_CONFIGURE          = "configure"
	ENDPOINT_CONF_FILE          = "config-file"
	ENDPOINT_SHOW_CONFIGURATION = "retrieve"

	ACTION_SET                 = "set"
	ACTION_DELETE              = "delete"
	ACTTION_SHOW_CONFIGURATION = "showConfig"

	VERSION_IPV4 = 4
	VERSION_IPV6 = 6
)

type Action struct {
	Op   string   `json:"op"`
	Path []string `json:"path"`
}

// 一次执行一条动作
func (c *Client) executeAction(endpoint string, action Action) ([]byte, error) {
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}
	return c.execute(endpoint, actionBytes)
}

// 一次执行多条动作
func (c *Client) executeBatchAction(endpoint string, actions []Action) ([]byte, error) {
	actionBytes, err := json.Marshal(actions)
	if err != nil {
		return nil, err
	}
	return c.execute(endpoint, actionBytes)
}

func (c *Client) execute(endpoint string, actionBytes []byte) ([]byte, error) {
	data := url.Values{}
	data.Add("data", string(actionBytes))
	data.Add("key", "")

	contentType := "application/x-www-form-urlencoded"
	return c.post(endpoint, nil, []byte(data.Encode()), contentType)
}

func (c *Client) decodeResp(resp []byte) error {
	var respInfo apiResponse
	if err := json.Unmarshal(resp, &respInfo); err != nil {
		return err
	}
	if !respInfo.Success {
		return fmt.Errorf("%s", respInfo.Error)
	}
	return nil
}

// SetAddress 设置网卡MTU和地址，address应由ip和掩码组成
func (c *Client) SetAddress(interfaceName, address string) error {
	var actions []Action
	// 网卡mtu必须设置
	actions = append(actions, Action{
		Op:   ACTION_SET,
		Path: []string{"interfaces", "ethernet", interfaceName, "mtu", "1450"},
	})
	// 地址不为空时再设置地址
	if len(address) != 0 {
		if _, _, err := net.ParseCIDR(address); err != nil {
			return fmt.Errorf("invalid address, address ust be in CIDR format, e.g. 192.168.1.1/24")
		}
		actions = append(actions, Action{
			Op:   ACTION_SET,
			Path: []string{"interfaces", "ethernet", interfaceName, "address", address},
		})
	}

	resp, err := c.executeBatchAction(ENDPOINT_CONFIGURE, actions)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// DeleteAddress 删除网卡地址，如果不指定具体地址，则直接删除指定网卡上的所有地址
func (c *Client) DeleteAddress(interfaceName, address string) error {
	action := Action{
		Op:   ACTION_DELETE,
		Path: []string{"interfaces", "ethernet", interfaceName, "address"},
	}
	if len(address) != 0 {
		if _, _, err := net.ParseCIDR(address); err != nil {
			return fmt.Errorf("invalid address, address ust be in CIDR format, e.g. 192.168.1.1/24")
		}
		action.Path = append(action.Path, address)
	}

	resp, err := c.executeAction(ENDPOINT_CONFIGURE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// DeleteInterface 删除网卡
func (c *Client) DeleteInterface(interfaceName string) error {
	action := Action{
		Op:   ACTION_DELETE,
		Path: []string{"interfaces", "ethernet", interfaceName},
	}
	resp, err := c.executeAction(ENDPOINT_CONFIGURE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// SaveConfig 持久化保存配置，任何对配置的修改最后都应调这个方法
func (c *Client) SaveConfig() error {
	action := Action{
		Op: "save",
	}

	resp, err := c.executeAction(ENDPOINT_CONF_FILE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// AddSnat 创建snat规则
func (c *Client) AddSnat(ruleId int, srcIp, translationIp, interfaceName string) error {
	srcIpInfo := net.ParseIP(srcIp)
	if srcIpInfo == nil {
		return fmt.Errorf("invalid srcIps %s", srcIp)
	}
	translationIpInfo := net.ParseIP(translationIp)
	if translationIpInfo == nil {
		return fmt.Errorf("invalid translationIp %s", translationIp)
	}
	// 判断两个ip的版本是否一致
	if (srcIpInfo.To4() != nil && translationIpInfo.To4() == nil) || (srcIpInfo.To4() == nil && translationIpInfo.To4() != nil) {
		return fmt.Errorf("ip类型不一致")
	}
	actions := []Action{
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "source", "rule", fmt.Sprintf("%d", ruleId), "outbound-interface", interfaceName},
		},
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "source", "rule", fmt.Sprintf("%d", ruleId), "source", "address", srcIp},
		},
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "source", "rule", fmt.Sprintf("%d", ruleId), "translation", "address", translationIp},
		},
	}

	resp, err := c.executeBatchAction(ENDPOINT_CONFIGURE, actions)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// DeleteSnat 删除snat
func (c *Client) DeleteSnat(ruleId int) error {
	action := Action{
		Op:   ACTION_DELETE,
		Path: []string{"nat", "source", "rule", fmt.Sprintf("%d", ruleId)},
	}
	resp, err := c.executeAction(ENDPOINT_CONFIGURE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// AddDnat 创建dnat规则
func (c *Client) AddDnat(ruleId int, dstIp, translationIp, interfaceName string) error {
	dstIpInfo := net.ParseIP(dstIp)
	if dstIpInfo == nil {
		return fmt.Errorf("invalid dstIP %s", dstIp)
	}
	translationIpInfo := net.ParseIP(translationIp)
	if translationIpInfo == nil {
		return fmt.Errorf("invalid translationIp %s", translationIp)
	}
	// 判断两个ip的版本是否一致
	if (dstIpInfo.To4() != nil && translationIpInfo.To4() == nil) || (dstIpInfo.To4() == nil && translationIpInfo.To4() != nil) {
		return fmt.Errorf("ip类型不一致")
	}

	actions := []Action{
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "destination", "rule", fmt.Sprintf("%d", ruleId), "inbound-interface", interfaceName},
		},
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "destination", "rule", fmt.Sprintf("%d", ruleId), "destination", "address", dstIp},
		},
		{
			Op:   ACTION_SET,
			Path: []string{"nat", "destination", "rule", fmt.Sprintf("%d", ruleId), "translation", "address", translationIp},
		},
	}

	resp, err := c.executeBatchAction(ENDPOINT_CONFIGURE, actions)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// DeleteDnat 删除dnat
func (c *Client) DeleteDnat(ruleId int) error {
	action := Action{
		Op:   ACTION_DELETE,
		Path: []string{"nat", "destination", "rule", fmt.Sprintf("%d", ruleId)},
	}
	resp, err := c.executeAction(ENDPOINT_CONFIGURE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)
}

// AddRoute 添加路由
func (c *Client) AddRoute(destNet, nextHop string, version int) error {
	// 校验destNat，提取出cidr，如192.168.1.1/24提取出192.168.1.0/24
	destNetIp, destCidr, err := net.ParseCIDR(destNet)
	if err != nil {
		return fmt.Errorf("invalid destNet %s", destNet)
	}

	nextHopIp := net.ParseIP(nextHop)
	if nextHopIp == nil {
		return fmt.Errorf("invalid nextHop %s", nextHop)
	}

	if version == VERSION_IPV4 && (destNetIp.To4() == nil || nextHopIp.To4() == nil) {
		return fmt.Errorf("ip类型不一致")
	}
	if version == VERSION_IPV6 && (destNetIp.To4() != nil || nextHopIp.To4() != nil) {
		return fmt.Errorf("ip类型不一致")
	}

	var routeType = "route"
	if version == VERSION_IPV6 {
		routeType = "route6"
	}
	action := Action{
		Op:   ACTION_SET,
		Path: []string{"protocols", "static", routeType, destCidr.String(), "next-hop", nextHop},
	}
	resp, err := c.executeAction(ENDPOINT_CONFIGURE, action)
	if err != nil {
		return err
	}
	return c.decodeResp(resp)

}

// ShowConfiguration 获取所有配置
func (c *Client) ShowConfiguration() (any, error) {
	action := Action{
		Op:   ACTTION_SHOW_CONFIGURATION,
		Path: []string{},
	}
	resp, err := c.executeAction(ENDPOINT_SHOW_CONFIGURATION, action)
	if err != nil {
		return nil, err
	}
	var respInfo apiResponse
	if err = json.Unmarshal(resp, &respInfo); err != nil {
		return nil, err
	}
	if !respInfo.Success {
		return nil, fmt.Errorf("%s", respInfo.Error)
	}
	return respInfo.Data, nil
}
