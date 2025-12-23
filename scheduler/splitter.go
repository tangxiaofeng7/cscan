package scheduler

import (
	"net"
	"strings"
)

// TargetSplitter 目标拆分器
type TargetSplitter struct {
	batchSize int // 每批次的IP数量
}

// NewTargetSplitter 创建目标拆分器
func NewTargetSplitter(batchSize int) *TargetSplitter {
	if batchSize <= 0 {
		batchSize = 50 // 默认每批50个IP
	}
	return &TargetSplitter{batchSize: batchSize}
}

// SplitTargets 拆分目标为多个批次
// 返回拆分后的目标列表，每个元素是一批目标（换行分隔）
func (s *TargetSplitter) SplitTargets(target string) []string {
	// 解析所有目标
	allTargets := s.parseAllTargets(target)

	// 如果目标数量小于等于批次大小，不拆分
	if len(allTargets) <= s.batchSize {
		return []string{target}
	}

	// 拆分为多个批次
	var batches []string
	for i := 0; i < len(allTargets); i += s.batchSize {
		end := i + s.batchSize
		if end > len(allTargets) {
			end = len(allTargets)
		}
		batch := strings.Join(allTargets[i:end], "\n")
		batches = append(batches, batch)
	}

	return batches
}

// parseAllTargets 解析所有目标（展开CIDR和IP范围）
func (s *TargetSplitter) parseAllTargets(target string) []string {
	var targets []string
	lines := strings.Split(target, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// CIDR格式
		if strings.Contains(line, "/") {
			ips := s.expandCIDR(line)
			targets = append(targets, ips...)
		} else if strings.Contains(line, "-") && net.ParseIP(strings.Split(line, "-")[0]) != nil {
			// IP范围格式 (确保是IP范围而不是域名中的连字符)
			ips := s.expandIPRange(line)
			targets = append(targets, ips...)
		} else {
			// 单个IP或域名
			targets = append(targets, line)
		}
	}

	return targets
}

// expandCIDR 展开CIDR
func (s *TargetSplitter) expandCIDR(cidr string) []string {
	var ips []string
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []string{cidr} // 解析失败，返回原始值
	}

	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); s.incIP(ip) {
		ips = append(ips, ip.String())
	}

	// 移除网络地址和广播地址
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}
	return ips
}

// expandIPRange 展开IP范围
func (s *TargetSplitter) expandIPRange(ipRange string) []string {
	var ips []string
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return []string{ipRange}
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))
	if startIP == nil || endIP == nil {
		return []string{ipRange}
	}

	// 复制起始IP，避免修改原始值
	ip := make(net.IP, len(startIP))
	copy(ip, startIP)

	for ; !ip.Equal(endIP); s.incIP(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, endIP.String())

	return ips
}

// incIP IP自增
func (s *TargetSplitter) incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// GetTargetCount 获取目标总数（不展开）
func (s *TargetSplitter) GetTargetCount(target string) int {
	return len(s.parseAllTargets(target))
}

// NeedSplit 判断是否需要拆分
func (s *TargetSplitter) NeedSplit(target string) bool {
	return s.GetTargetCount(target) > s.batchSize
}
