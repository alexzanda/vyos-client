package vyos

// 基于http api的vyos客户端实现，已建单实现针对网卡、snat、dnat相关的配置实现，vyos的配置需要持久化保存到其配置文件才能避免在重启的时候
// 配置丢失，虽然模板中已经配置自动定时保存配置，但是为了安全起见，每次通过api修改完配置后，都应调用SaveConfig进行手动触发保存
