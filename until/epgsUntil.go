package until

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"log"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ConvertCntvToXml(cntv dto.CntvJsonChannel, eName string) dto.XmlTV {
	tv := dto.XmlTV{
		GeneratorName: "清和IPTV管理系统",
		GeneratorURL:  "https://www.qingh.xyz",
	}

	// 添加频道
	tv.Channels = append(tv.Channels, dto.XmlChannel{
		ID: eName,
		DisplayName: []dto.DisplayName{
			{Lang: "zh",
				Value: eName,
			},
		},
	})

	// 添加节目表
	for _, p := range cntv.Program {
		start := time.Unix(p.StartTime, 0).UTC().Format("20060102150405 -0700")
		stop := time.Unix(p.EndTime, 0).UTC().Format("20060102150405 -0700")

		tv.Programmes = append(tv.Programmes, dto.Programme{
			Start:   start,
			Stop:    stop,
			Channel: eName,
			Title: dto.Title{
				Lang:  "zh",
				Value: p.Title,
			},
			Desc: dto.Desc{
				Lang:  "zh",
				Value: p.Title,
			},
		})
	}

	return tv
}

func GetEpgListXml(name, url string) dto.XmlTV {
	epgUrl := url
	cacheKey := "epgXmlFrom_" + name
	var xmlTV dto.XmlTV
	var xmlByte []byte
	readCacheOk := false
	if dao.Cache.Exists(cacheKey) {
		tmpByte, err := dao.Cache.Get(cacheKey)
		if err == nil {
			xmlByte = tmpByte
			readCacheOk = true
		}
	}

	if !readCacheOk {
		xmlByte = []byte(GetUrlData(epgUrl))
		if dao.Cache.Set(cacheKey, xmlByte) != nil {
			dao.Cache.Delete(cacheKey)
		}
	}
	xml.Unmarshal(xmlByte, &xmlTV)
	return xmlTV
}

func GetEpgCntv(name string) (dto.CntvJsonChannel, error) {

	var cacheKey = "cntv_" + strings.ToUpper(name)

	var cntvJson dto.CntvData

	if name == "" {
		return dto.CntvJsonChannel{}, errors.New("id is empty")
	}
	name = strings.ToLower(name)

	epgUrl := "https://api.cntv.cn/epg/epginfo?c=" + name + "&serviceId=channel&d="

	readCacheOk := false
	if dao.Cache.Exists(cacheKey) {
		err := dao.Cache.GetJSON(cacheKey, cntvJson)
		if err == nil {
			readCacheOk = true
		}
	}

	if !readCacheOk {
		jsonStr := GetUrlData(epgUrl)
		err := json.Unmarshal([]byte(jsonStr), &cntvJson)
		if err != nil {
			return dto.CntvJsonChannel{}, err
		}
		if dao.Cache.SetJSON(cacheKey, cntvJson) != nil {
			dao.Cache.Delete(cacheKey)
		}
	}
	return cntvJson[name], nil
}

func UpdataEpgList() bool {
	var epgLists []models.IptvEpgList
	dao.DB.Model(&models.IptvEpgList{}).Find(&epgLists)
	for _, list := range epgLists {
		log.Println("更新EPG源: ", list.Name)
		cacheKey := "epgXmlFrom_" + list.Name
		dao.Cache.Delete(cacheKey)
		xmlStr := GetUrlData(strings.TrimSpace(list.Url), list.UA)
		if xmlStr != "" {
			xmlByte := []byte(xmlStr)
			if dao.Cache.Set(cacheKey, xmlByte) != nil {
				dao.Cache.Delete(cacheKey)
			}
			var xmlTV dto.XmlTV
			if xml.Unmarshal(xmlByte, &xmlTV) != nil {
				continue
			}
			var epgs []models.IptvEpg
			// 1️⃣ 匹配数字台，如 CCTV1、CCTV-5+、CCTV13 等
			reNum := regexp.MustCompile(`(?i)CCTV-?(\d+\+?)$`)

			// 2️⃣ 匹配字母台，如 CCTV4EUO、CCTV4AME、CCTVF、CCTVE 等
			reAlpha := regexp.MustCompile(`(?i)CCTV(\d*[A-Z]+)`)
			for _, channel := range xmlTV.Channels {
				remarks := channel.DisplayName[0].Value
				upper := strings.ToUpper(remarks)
				if strings.Contains(upper, "CCTV") {
					switch {
					case reNum.MatchString(upper):
						match := reNum.FindStringSubmatch(upper)
						num := match[1]
						remarks = fmt.Sprintf("CCTV%s|CCTV-%s|CCTV%s 4K|CCTV-%s 4K|CCTV%s HD|CCTV-%s HD", num, num, num, num, num, num)

					case reAlpha.MatchString(upper):
						match := reAlpha.FindStringSubmatch(upper)
						suffix := match[1]
						remarks = fmt.Sprintf("CCTV%s|CCTV-%s", suffix, suffix)
					}
				} else {
					remarks = fmt.Sprintf("%s|%s 4K|%s HD", remarks, remarks, remarks)
				}
				epgs = append(epgs, models.IptvEpg{
					Name:    channel.DisplayName[0].Value,
					Status:  1,
					Remarks: remarks,
				})
			}
			if len(epgs) > 0 {
				dao.DB.Model(&models.IptvEpgList{}).Where("id = ?", list.ID).Updates(&models.IptvEpgList{Status: 1, LastTime: time.Now().Unix()})
				// dao.DB.Model(&models.IptvEpg{}).Where("name like ?", list.Remarks+"-%").Delete(&models.IptvEpg{})
				// dao.DB.Model(&models.IptvEpg{}).Create(&epgs)
				log.Println("开始同步EPG")
				reload, _ := SyncEpgs(list.ID, epgs, false) // 同步
				if reload {
					go BindChannel() // 绑定频道
				} else {
					go CleanMealsEpgCacheAll()
				}
			}
		}
	}
	log.Println("EPG列表更新完成")
	return true
}

func UpdataEpgListOne(list models.IptvEpgList, newAdd bool) (bool, error) {
	log.Println("更新EPG源: ", list.Name)
	cacheKey := "epgXmlFrom_" + list.Name
	dao.Cache.Delete(cacheKey)
	xmlStr := GetUrlData(strings.TrimSpace(list.Url), list.UA)
	if xmlStr != "" {
		xmlByte := []byte(xmlStr)
		if dao.Cache.Set(cacheKey, xmlByte) != nil {
			dao.Cache.Delete(cacheKey)
		}
		var xmlTV dto.XmlTV
		if xml.Unmarshal(xmlByte, &xmlTV) != nil {
			return false, errors.New("xml解析失败")
		}
		var epgs []models.IptvEpg
		// 1️⃣ 匹配数字台，如 CCTV1、CCTV-5+、CCTV13 等
		reNum := regexp.MustCompile(`(?i)CCTV-?(\d+\+?)$`)

		// 2️⃣ 匹配字母台，如 CCTV4EUO、CCTV4AME、CCTVF、CCTVE 等
		reAlpha := regexp.MustCompile(`(?i)CCTV(\d*[A-Z]+)`)
		for _, channel := range xmlTV.Channels {
			remarks := channel.DisplayName[0].Value
			if remarks == "" {
				continue
			}
			upper := strings.ToUpper(remarks)
			if strings.Contains(upper, "CCTV") {
				switch {
				case reNum.MatchString(upper):
					match := reNum.FindStringSubmatch(upper)
					num := match[1]
					remarks = fmt.Sprintf("CCTV%s|CCTV-%s|CCTV%s 4K|CCTV-%s 4K|CCTV%s HD|CCTV-%s HD", num, num, num, num, num, num)

				case reAlpha.MatchString(upper):
					match := reAlpha.FindStringSubmatch(upper)
					suffix := match[1]
					remarks = fmt.Sprintf("CCTV%s|CCTV-%s", suffix, suffix)
				}
			} else {
				remarks = fmt.Sprintf("%s|%s 4K|%s HD", remarks, remarks, remarks)
			}

			epgs = append(epgs, models.IptvEpg{
				Name:    channel.DisplayName[0].Value,
				Status:  1,
				Remarks: remarks,
			})
		}
		if len(epgs) > 0 {
			dao.DB.Model(&models.IptvEpgList{}).Where("id = ?", list.ID).Updates(&models.IptvEpgList{Status: 1, LastTime: time.Now().Unix()})
			// dao.DB.Model(&models.IptvEpg{}).Where("name like ?", list.Remarks+"-%").Delete(&models.IptvEpg{})
			// dao.DB.Model(&models.IptvEpg{}).Create(&epgs)

			log.Println("开始同步EPG")
			reload, _ := SyncEpgs(list.ID, epgs, newAdd) // 同步
			if reload {
				go BindChannel() // 绑定频道
			} else {
				go CleanMealsEpgCacheAll()
			}

			log.Println("EPG更新完成")
			return true, nil
		}
		return false, errors.New("未找到epg数据")
	}
	return false, errors.New("URL错误:" + list.Url)
}

func BindChannel() bool {
	// ClearBind() // 清空绑定

	var epgList []models.IptvEpg
	if err := dao.DB.Model(&models.IptvEpg{}).Where("status = 1").Find(&epgList).Error; err != nil {
		return false
	}
	channelCache := make(map[string][]models.IptvChannel)

	var update = false
	var upCaList []string
	for _, epgData := range epgList {
		if epgData.CasStr == "" {
			continue
		}
		caList := strings.Split(epgData.CasStr, ",")
		if len(caList) == 0 || (len(caList) == 1 && caList[0] == "") {
			continue
		}

		cacheKey := getCAKey(caList)

		var channelList []models.IptvChannel
		if val, ok := channelCache[cacheKey]; ok {
			channelList = val
		} else {
			dao.DB.Model(&models.IptvChannel{}).
				Select("distinct name").
				Where("status = 1 and c_id in (?)", caList).
				Find(&channelList)
			channelCache[cacheKey] = channelList
		}
		var tmpList []string
		nameList := strings.Split(epgData.Remarks, "|")

		for _, channelData := range channelList {
			if strings.EqualFold(channelData.Name, epgData.Name) {
				tmpList = append(tmpList, channelData.Name)
				continue
			}

			for _, name := range nameList {
				if strings.EqualFold(channelData.Name, name) || channelData.Name == name {
					tmpList = append(tmpList, channelData.Name)
					break
				}
			}
		}
		chNameList := MergeAndUnique(strings.Split(epgData.Content, ","), tmpList)

		if len(tmpList) > 0 {
			update = true
			dao.DB.Model(&models.IptvChannel{}).Where("name in (?) and c_id in (?) and status = 1", chNameList, caList).Update("e_id", epgData.ID)

			upCaList = MergeAndUnique(upCaList, caList) // 记录需要更新的分类ID列表

			if !EqualStringSets(strings.Split(epgData.Content, ","), chNameList) {
				epgData.Content = strings.Join(chNameList, ",")
				if epgData.Content != "" {
					dao.DB.Save(&epgData)
				}
			}
		}
	}

	if update {
		go checkCaIdsInMeals(upCaList)
		cfg := dao.GetConfig()
		if cfg.Epg.Fuzz == 1 && dao.Lic.Type != 0 {
			dao.WS.SendWS(dao.Request{Action: "checkChEpg"})
			CleanMealsEpgCacheAll()
		}
	}
	return true
}

func checkCaIdsInMeals(ids []string) {
	var rebuild = false
	for _, id := range ids {
		var count int64
		err := dao.DB.Model(&models.IptvMeals{}).
			Where("(content = ? OR content LIKE ? OR content LIKE ? OR content LIKE ?) AND status = 1",
				id,           // 单独一个值
				id+",%",      // 开头
				"%,"+id+",%", // 中间
				"%,"+id,      // 结尾
			).
			Count(&count).Error
		if err != nil {
			continue
		}
		if count > 0 {
			rebuild = true
			break
		}
	}

	if rebuild {
		CleanMealsRssCacheAll()
	}
}

func getCAKey(caList []string) string {
	var intList []int
	for _, s := range caList {
		if s == "" {
			continue
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		intList = append(intList, n)
	}
	sort.Ints(intList)

	var strList []string
	for _, n := range intList {
		strList = append(strList, strconv.Itoa(n))
	}
	return strings.Join(strList, ",")
}

// SyncEpgs 同步 IPTV EPG 数据：
// - 保留数据库中已存在的记录（不更新）
// - 新数据中有但数据库没有的 → 新增
// - 数据库中有但新数据中没有的 → 删除
func SyncEpgs(fromId int64, epgs []models.IptvEpg, newAdd bool) (bool, error) {
	// 1. 查询数据库中已有的记录
	var oldEpgs []models.IptvEpg
	if err := dao.DB.Model(&models.IptvEpg{}).Where("status = 1").Find(&oldEpgs).Error; err != nil {
		return false, err
	}

	// 2. 建立 name 映射方便比对
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)
	for _, o := range oldEpgs {
		oldMap[o.Name] = true
		for i, n := range epgs {
			if o.Name == n.Name {
				epgs[i].ID = o.ID
				epgs[i].FromListStr = o.FromListStr
				epgs[i].Content = o.Content
				epgs[i].Remarks = o.Remarks
				epgs[i].Status = o.Status
				epgs[i].CasStr = o.CasStr
			}
			newMap[n.Name] = true
		}
	}

	// 3. 计算需要新增与删除的数据
	var toAdd []models.IptvEpg

	for _, n := range epgs {
		if !oldMap[n.Name] || newAdd {
			toAdd = append(toAdd, n)
		}
	}

	for _, o := range oldEpgs {
		if !newMap[o.Name] {
			tmpList := strings.Split(o.FromListStr, ",")
			exist := false
			for i, v := range tmpList {
				if v == fmt.Sprintf("%d", fromId) {
					exist = true
					tmpList = append(tmpList[:i], tmpList[i+1:]...)
					break // 若只删除第一个匹配项
				}
			}

			if exist {
				tmpList = RemoveEmptyStrings(tmpList)
				if len(tmpList) > 0 {
					dao.DB.Model(&models.IptvEpg{}).Where("id = ?", o.ID).Update("fromlist", strings.Join(tmpList, ","))
				}
			}
		}
	}
	addCount := 0
	if len(toAdd) > 0 {
		var caIDs []int64
		dao.DB.Model(&models.IptvCategory{}).
			Where("enable = 1 AND type not like ?", "auto%").
			Pluck("id", &caIDs)

		for _, toAddOne := range toAdd {
			oldList := strings.Split(toAddOne.FromListStr, ",")
			tmpList := append(oldList, fmt.Sprintf("%d", fromId))
			tmpList = RemoveEmptyStrings(tmpList)
			toAddOne.FromListStr = strings.Join(tmpList, ",")

			if EqualStringSets(oldList, tmpList) {
				continue
			}
			if toAddOne.ID == 0 {
				toAddOne.CasStr = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(caIDs)), ","), "[]") // 转换为字符串
			}
			addCount++
			dao.DB.Save(&toAddOne)
		}
		log.Printf("新增 %d 条 EPG 记录\n", addCount)
	}
	if addCount > 0 {
		return true, nil
	}
	return false, errors.New("无新增数据")
}

func GetEpg(id int64) dto.XmlTV {

	res := dto.XmlTV{
		GeneratorName: "清和IPTV管理系统",
		GeneratorURL:  "https://www.qingh.xyz",
	}

	epgCaCheKey := "rssEpgXml_" + strconv.FormatInt(id, 10)
	if dao.Cache.Exists(epgCaCheKey) {
		cacheData, err := dao.Cache.Get(epgCaCheKey)
		if err == nil {
			err := xml.Unmarshal(cacheData, &res)
			if err == nil {
				return res
			}
		}
	}

	var meal models.IptvMeals
	if err := dao.DB.Model(&models.IptvMeals{}).Where("id = ? and status = 1", id).First(&meal).Error; err != nil {
		return res
	}
	raw := strings.Split(meal.Content, ",")
	categoryIdList := make([]string, 0, len(raw))
	for _, s := range raw {
		if s != "" {
			categoryIdList = append(categoryIdList, s)
		}
	}
	if len(categoryIdList) == 0 {
		return res
	}
	var categoryList []models.IptvCategory
	if err := dao.DB.Model(&models.IptvCategory{}).Where("id in (?) and enable = 1", categoryIdList).Order("sort asc").Find(&categoryList).Error; err != nil {
		return res
	}

	var channels []models.IptvChannelShow
	for _, category := range categoryList {
		if strings.Contains(category.Type, "auto") {
			channels = append(channels, GetAutoChannelList(category, false)...)
		} else {
			var tmpChannels []models.IptvChannelShow
			dao.DB.Model(&models.IptvChannelShow{}).Where("c_id = ? and status = 1", category.ID).Order("sort asc").Find(&tmpChannels)
			channels = append(channels, tmpChannels...)
		}
	}

	res = GetEpgXml(channels)
	CleanTV(&res)

	data, err := xml.Marshal(res)
	if err == nil {
		err := dao.Cache.Set(epgCaCheKey, data)
		if err != nil {
			log.Println("epg缓存设置失败:", err)
			dao.Cache.Delete(epgCaCheKey)
		}
	} else {
		log.Println("epg缓存序列化失败:", err)
		dao.Cache.Delete(epgCaCheKey)
	}
	return res
}

func CleanTV(tv *dto.XmlTV) {
	// ===== Channel 去重 + ID 重映射 =====
	chLen := len(tv.Channels)
	newChannels := make([]dto.XmlChannel, 0, chLen)

	seen := make(map[string]struct{}, chLen)
	idMap := make(map[string]string, chLen)

	nextID := 1
	for i := range tv.Channels {
		ch := &tv.Channels[i]
		if _, ok := seen[ch.ID]; ok {
			continue
		}
		seen[ch.ID] = struct{}{}

		newID := strconv.Itoa(nextID)
		idMap[ch.ID] = newID
		ch.ID = newID

		newChannels = append(newChannels, *ch)
		nextID++
	}
	tv.Channels = newChannels

	// ===== Programme 确定性去重 =====
	progLen := len(tv.Programmes)
	newProgrammes := make([]dto.Programme, 0, progLen)
	progIndex := make(map[string]int, progLen)

	for i := range tv.Programmes {
		p := &tv.Programmes[i]

		newCID, ok := idMap[p.Channel]
		if !ok {
			continue
		}
		p.Channel = newCID

		startKey := p.Start
		if len(startKey) >= 14 {
			startKey = startKey[:14]
		}

		key := newCID + "_" + startKey + "_" + p.Title.Value

		if idx, exists := progIndex[key]; exists {
			old := &newProgrammes[idx]
			oldHasTZ := strings.Contains(old.Start, "+") || strings.Contains(old.Start, "-")
			newHasTZ := strings.Contains(p.Start, "+") || strings.Contains(p.Start, "-")

			if !oldHasTZ && newHasTZ {
				newProgrammes[idx] = *p
			}
			continue
		}

		progIndex[key] = len(newProgrammes)
		newProgrammes = append(newProgrammes, *p)
	}

	tv.Programmes = newProgrammes
}

func GetEpgXml(channelList []models.IptvChannelShow) dto.XmlTV {
	epgXml := dto.XmlTV{
		GeneratorName: "清和IPTV管理系统",
		GeneratorURL:  "https://www.qingh.xyz",
	}

	// ===== 核心缓存 =====
	epgXmlExist := make(map[string]struct{})            // channel.Name 是否已生成
	epgCache := make(map[int64]models.IptvEpg)          // IptvEpg 表缓存
	epgListCache := make(map[string]models.IptvEpgList) // IptvEpgList 表缓存
	channelIndex := make(map[string]int)                // epg.Name -> Channels index
	epgXmlCache := make(map[string]*dto.XmlTV)          // 零重复解析缓存

	for _, channel := range channelList {
		if channel.EId <= 0 {
			continue
		}
		if _, ok := epgXmlExist[channel.Name]; ok {
			continue
		}

		// ===== 获取 EPG =====
		epg, ok := epgCache[channel.EId]
		if !ok {
			var tmp models.IptvEpg
			if err := dao.DB.Where("id = ? and status = 1", channel.EId).First(&tmp).Error; err != nil {
				continue
			}
			epgCache[channel.EId] = tmp
			epg = tmp
		}

		fromList := strings.Split(epg.FromListStr, ",")
		if len(fromList) == 0 {
			continue
		}

		// ===== CNTV 优先 =====
		if slices.Contains(fromList, "0") {
			name := epg.Name
			if strings.EqualFold(name, "cctv5+") || strings.EqualFold(name, "cctv-5+") {
				name = "cctv5plus"
			}

			if tmpData, err := GetEpgCntv(name); err == nil {
				tmpXml := ConvertCntvToXml(tmpData, name)

				mergeChannel(&epgXml, channelIndex, name, channel.Name)

				for _, p := range tmpXml.Programmes {
					p2 := p
					p2.Channel = name
					epgXml.Programmes = append(epgXml.Programmes, p2)
				}

				epgXmlExist[channel.Name] = struct{}{}
				continue
			}
		}

		// ===== 其他来源 =====
		for _, from := range fromList {
			if from == "" || from == "0" {
				continue
			}

			epgFrom, ok := epgListCache[from]
			if !ok {
				var tmp models.IptvEpgList
				if err := dao.DB.Where("id = ? and status = 1", from).First(&tmp).Error; err != nil {
					continue
				}
				epgListCache[from] = tmp
				epgFrom = tmp
			}

			if epgFrom.Url == "" || epgFrom.Name == "" {
				continue
			}

			tmpXml := getEpgListXmlCached(epgFrom.Name, epgFrom.Url, epgXmlCache)

			mergeChannel(&epgXml, channelIndex, epg.Name, channel.Name)

			var srcID string
			for _, c := range tmpXml.Channels {
				if len(c.DisplayName) > 0 && c.DisplayName[0].Value == epg.Name {
					srcID = c.ID
					break
				}
			}

			for _, p := range tmpXml.Programmes {
				if p.Channel == srcID {
					p2 := p
					p2.Channel = epg.Name
					epgXml.Programmes = append(epgXml.Programmes, p2)
				}
			}

			epgXmlExist[channel.Name] = struct{}{}
			break
		}
	}

	return epgXml
}

func mergeChannel(
	epgXml *dto.XmlTV,
	index map[string]int,
	channelID, displayName string,
) {
	if i, ok := index[channelID]; ok {
		for _, d := range epgXml.Channels[i].DisplayName {
			if d.Value == displayName {
				return
			}
		}
		epgXml.Channels[i].DisplayName = append(
			epgXml.Channels[i].DisplayName,
			dto.DisplayName{Lang: "zh", Value: displayName},
		)
		return
	}

	index[channelID] = len(epgXml.Channels)
	epgXml.Channels = append(epgXml.Channels, dto.XmlChannel{
		ID: channelID,
		DisplayName: []dto.DisplayName{
			{Lang: "zh", Value: displayName},
		},
	})
}

func getEpgListXmlCached(
	name, url string,
	cache map[string]*dto.XmlTV,
) *dto.XmlTV {

	key := name + "|" + url
	if tv, ok := cache[key]; ok {
		return tv
	}

	tv := GetEpgListXml(name, url)
	cache[key] = &tv
	return &tv
}
