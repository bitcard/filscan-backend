package controllers

import (
	"context"
	"encoding/json"
	"errors"
	filscanproto "filscan_lotus/filscanproto"
	"filscan_lotus/models"
	"gitlab.forceup.in/dev-proto/common"
	"strconv"
)

var _ filscanproto.FilscanServer = (*FilscanServer)(nil)

type FilscanServer struct {
}

/**
SearchIndex(context.Context, *SearchIndexReq) (*SearchIndexResp, error)
BaseInformation(context.Context, *Empty) (*BaseInformationResp, error)
BlocktimeGraphical(context.Context, *Empty) (*BlocktimeGraphicalResp, error)
TotalPowerGraphical(context.Context, *Empty) (*TotalPowerGraphicalResp, error)
FILOutsanding(context.Context, *Empty) (*FILOutsandingResp, error)
LatestBlock(context.Context, *LatestBlockReq) (*LatestBlockResp, error)
LatestMsg(context.Context, *LatestMsgReq) (*LatestMsgResp, error)
*/

func (this *FilscanServer) SearchIndex(ctx context.Context, input *filscanproto.SearchIndexReq) (*filscanproto.SearchIndexResp, error) {
	resp := new(filscanproto.SearchIndexResp)
	key := input.GetKey()
	filter := input.GetFilter() // 1-address , 2-message_ID , 3-Height , 4-block_hash , 5-peer_id
	if !CheckArg(key) {
		res := &common.Result{Code: 5, Msg: "Missing required parameters"}
		resp.Res = res
		return resp, nil
	}

	res := "" // actor ,message_ID ,Height , block_hash , peer_id
	msg := "sucess"
	switch filter {
	case 1:
		//msg = "Unrealized"
		goto SearchActor
	case 2:
		goto SearchMsg
	case 3:
		goto Search_Height
	case 4:
		goto Search_Block_Hash
	case 5:
		//msg = "Unrealized"
		goto SearchPeer
	default:
		goto SearchActor
		if len(res) > 0 {
			break
		}
		goto SearchMsg
		goto Search_Height
		goto Search_Block_Hash
	}
SearchActor:
	{
		actor, err := models.GetActorByAddress(key)
		if err != nil {
			log("SearchIndex GetActorByAddress search err=%v", err)
			res := &common.Result{Code: 5, Msg: "SearchIndex GetActorByAddress search err"}
			resp.Res = res
			return resp, nil
		}
		if actor != nil {
			res = "actor"
		}
	}
SearchMsg:
	{
		msg, err := models.GetMsgByMsgCid(key)
		if err != nil {
			log("SearchIndex GetMsgByMsgCid search err=%v", err)
			res := &common.Result{Code: 5, Msg: "SearchIndex GetMsgByMsgCid search err"}
			resp.Res = res
			return resp, nil
		}
		if len(msg) > 0 {
			res = "message_ID"
		} else {
			m := TipsetQueue.MsgByCid(key)
			if m != nil {
				res = "message_ID"
			}
		}
	}
Search_Height:
	{
		height, _ := strconv.ParseUint(key, 10, 64)
		//if height > 0 && TipsetQueue.Size() > 0 && TipsetQueue.element[TipsetQueue.Size()-1].tipset.Height() > height {
		t, err := models.GetMaxTipSet()
		if err != nil {
			log("SearchIndex GetTipSetByOneHeight search err=%v", err)
			res := &common.Result{Code: 5, Msg: "SearchIndex GetTipSetByOneHeight search err"}
			resp.Res = res
			return resp, nil
		}
		if (height > 0 && height < t.Height) || (height > 0 && TipsetQueue.Size() > 0 && TipsetQueue.element[TipsetQueue.Size()-1].tipset.Height() > height) {
			res = "Height"
		} else {
			e := TipsetQueue.TipsetByOneHeight(height)
			if e != nil {
				res = "Height"
			}
		}
	}
Search_Block_Hash:
	{
		cid := []string{}
		cid = append(cid, key)
		bs, err := models.GetBlockByCid(cid)
		if err != nil {
			log("SearchIndex GetBlockByCid search err=%v", err)
			res := &common.Result{Code: 5, Msg: "SearchIndex GetBlockByCid search err"}
			resp.Res = res
			return resp, nil
		}
		if len(bs) > 0 {
			res = "block_hash"
		} else {
			block := TipsetQueue.BlockByCid(key)
			if block != nil {
				res = "block_hash"
			}
		}
	}
SearchPeer:
	{
		peer, err := models.MinerByPeerId(key)
		if err != nil {
			log("SearchIndex MinerByPeerId search err=%v", err)
			res := &common.Result{Code: 5, Msg: "SearchIndex MinerByPeerId search err"}
			resp.Res = res
			return resp, nil
		}
		if peer != nil {
			res = "peer_id"
		}
	}

	resp.Res = common.NewResult(3, msg)
	resp.Data = &filscanproto.SearchIndexResp_Data{ModelFlag: res}
	return resp, nil
}

func (this *FilscanServer) BaseInformation(ctx context.Context, input *common.Empty) (*filscanproto.BaseInformationResp, error) {
	resp := new(filscanproto.BaseInformationResp)

	if models.TimeNow-lotusBaseInformation.Time > lotusBaseInformation.CashTime {
		totalPrice, err := models.GetSumGasPriceByMsgMinCreat(models.TimeNow - 60*60*24)
		if err != nil {
			log("GetSumGasPrice search err=%v", err)
			resp.Res = &common.Result{Code: 5, Msg: "GetSumGasPrice search err"}
			return resp, nil
		}
		totalSize, err := models.GetSumSizeByMsgMinCreat(models.TimeNow - 60*60*24)
		if err != nil {
			log("GetSumSize search err=%v", err)
			resp.Res = &common.Result{Code: 5, Msg: "GetSumSize search err"}
			return resp, nil
		}
		msgCount, err := models.GetMsgCountByMsgMinCreat(models.TimeNow - 60*60*24)
		if err != nil {
			log("GetMsgCount search err=%v", err)
			resp.Res = &common.Result{Code: 5, Msg: "GetMsgCount search err"}
			return resp, nil
		}
		tipsetCount, err := models.GetTipsetCountByMinCreat(models.TimeNow - 60*60*24)
		if err != nil {
			log("GetTipsetCount search err=%v", err)
			resp.Res = &common.Result{Code: 5, Msg: "GetTipsetCount search err"}
			return resp, nil
		}
		pcString := ""
		tipset, err := GetLotusHead()
		if err != nil {
			log("GetLotusHead search err=%v", err)
			//resp.Res  = &common.Result{Code: 5, Msg: "GetLotusHead err "}
			//resp.Res = res
			//return resp, nil
		} else {
			pcString, err = GetPledgeCollateral(tipset)
			if err != nil {
				log("GetPledgeCollateral search err=%v", err)
				resp.Res = &common.Result{Code: 5, Msg: "GetPledgeCollateral err "}

				return resp, nil
			}
		}
		if TipsetQueue.Size() > 0 {
			lotusBaseInformation.TipsetHeight = TipsetQueue.element[len(TipsetQueue.element)-1].tipset.Height()
		} else {
			lotusBaseInformation.TipsetHeight = tipset.Height()
			if tipset.Height() < 1 {
				res, _ := models.GetMaxTipSet()
				lotusBaseInformation.TipsetHeight = res.Height
			}
		}
		allMsg := TipsetQueue.AllMsg()
		var cashTotalPrice, cashTotalSize uint64
		for _, value := range allMsg {
			cashTotalPrice += value.Message.GasPrice.Uint64()
			cashTotalSize += uint64(value.Size)
		}
		if msgCount+len(allMsg) > 0 {
			lotusBaseInformation.AvgGasPrice = (totalPrice + float64(cashTotalPrice)) / float64(msgCount+len(allMsg))
			count := msgCount + len(allMsg)
			lotusBaseInformation.AvgMessageSize = float64(totalSize+cashTotalSize) / float64(count)
		}
		if (tipsetCount + TipsetQueue.Size()) > 0 {
			lotusBaseInformation.AvgMessagesTipset = float64((msgCount + len(allMsg)) / (tipsetCount + TipsetQueue.Size()))
		}
		latestReward, err := models.GetLatestReward()
		if err != nil {
			log("GetLatestReward search err=%v", err)
			lotusBaseInformation.BlockReward = 0
		} else {
			lotusBaseInformation.BlockReward, _ = strconv.ParseFloat(latestReward, 64)
		}
		lotusBaseInformation.PledgeCollateral = pcString
		lotusBaseInformation.Time = models.TimeNow
	}
	var bim filscanproto.BaseInformationResp_Data
	bim.TipsetHeight = lotusBaseInformation.TipsetHeight
	bim.BlockReward = lotusBaseInformation.BlockReward
	bim.AvgMessageSize = lotusBaseInformation.AvgMessageSize
	bim.AvgGasPrice = lotusBaseInformation.AvgGasPrice
	bim.AvgMessagesTipset = lotusBaseInformation.AvgMessagesTipset
	bim.PledgeCollateral = lotusBaseInformation.PledgeCollateral
	resp.Data = &bim
	resp.Res = common.NewResult(3, "success")
	return resp, nil
}

func (this *FilscanServer) BlocktimeGraphical(ctx context.Context, input *filscanproto.StartEndTimeReq) (*filscanproto.BlocktimeGraphicalResp, error) {
	resp := new(filscanproto.BlocktimeGraphicalResp)
	startTime := input.GetStartTime()
	endTime := input.GetEndTime()
	if !CheckArg(startTime, endTime) {
		res := &common.Result{Code: 5, Msg: "Missing required parameters"}
		resp.Res = res
		return resp, nil
	}
	tArr, err := GetIntHour(startTime, endTime)

	if endTime-avgBlockTimeCash.Time <= avgBlockTimeCash.CashTime {
		resp.Data = &filscanproto.BlocktimeGraphicalResp_Data{Data: avgBlockTimeCash.BlockTime, AvgBlocktime: avgBlockTimeCash.AvgBlockTime,
			Max: avgBlockTimeCash.Max, Min: avgBlockTimeCash.Min}
		resp.Res = common.NewResult(3, "success")
		return resp, nil
	}
	if len(tArr) > 25 {
		res := &common.Result{Code: 5, Msg: "Length of Time Err"}
		resp.Res = res
		return resp, nil
	}
	if err != nil {
		log("search err=%v ", err)
		res := &common.Result{Code: 5, Msg: "search err"}
		resp.Res = res
		return resp, nil
	}
	var bts []*filscanproto.Blocktime
	var totalBlockCount int
	var min string
	var max string
	var firstIndex int
	var flag bool
	for key := range tArr {
		blocktime := new(filscanproto.Blocktime)
		start := tArr[key]
		var end int64
		if key == len(tArr)-1 {
			end = models.TimeNow
			continue
		} else {
			end = tArr[key+1]
		}
		bms, err := GetTipsetNumByTime(start, end)
		if err != nil {
			log("search err=%v ", err)
			res := &common.Result{Code: 5, Msg: "search err"}
			resp.Res = res
			return resp, nil
		}
		if bms == 0 {
			continue
		}
		if bms > 0 && !flag {
			firstIndex = key
		}
		totalBlockCount += bms
		blocktime.Time = start
		if bms > 0 {
			t := RoundString(float64((end-start))/float64(bms), 0, false)
			blocktime.BlockTime = t
			if !flag {
				max = t
				min = t
				flag = true
			} else {
				if t > max {
					max = t
				}
				if t < min {
					min = t
				}
			}
		} else {
			blocktime.BlockTime = "0"
		}
		bts = append(bts, blocktime)
	}
	avgBlockTimeCash.Min = min
	avgBlockTimeCash.Max = max
	avgBlockTimeCash.TotalBlockCount = totalBlockCount
	avgBlockTimeCash.Time = endTime
	avgBlockTimeCash.BlockTime = bts
	avgBlockTimeCash.AvgBlockTime = RoundString(float64(models.TimeNow-tArr[firstIndex])/float64(int64(totalBlockCount)), 0, false)
	resp.Data = &filscanproto.BlocktimeGraphicalResp_Data{Data: bts, AvgBlocktime: RoundString(float64(models.TimeNow-tArr[firstIndex])/float64(int64(totalBlockCount)), 0, false),
		Max: max, Min: min}
	resp.Res = common.NewResult(3, "success")
	return resp, nil
}

func (this *FilscanServer) AvgBlockheaderSizeGraphical(ctx context.Context, input *filscanproto.StartEndTimeReq) (*filscanproto.AvgBlockheaderSizeGraphicalResp, error) {
	resp := new(filscanproto.AvgBlockheaderSizeGraphicalResp)
	startTime := input.GetStartTime()
	endTime := input.GetEndTime()
	if !CheckArg(startTime, endTime) {
		res := &common.Result{Code: 5, Msg: "Missing required parameters"}
		resp.Res = res
		return resp, nil
	}
	tArr, err := GetIntHour(startTime, endTime)

	if endTime-avgBlockSizeCash.Time <= avgBlockSizeCash.CashTime {
		resp.Data = &filscanproto.AvgBlockheaderSizeGraphicalResp_Data{Data: avgBlockSizeCash.BlockSize, AvgBlocksize: avgBlockSizeCash.AvgBlocksize, Min: avgBlockSizeCash.Min, Max: avgBlockSizeCash.Max}
		resp.Res = common.NewResult(3, "success")
		return resp, nil
	}
	if len(tArr) > 25 {
		res := &common.Result{Code: 5, Msg: "Length of Time Err"}
		resp.Res = res
		return resp, nil
	}
	if err != nil {
		log("search err=%v ", err)
		res := &common.Result{Code: 5, Msg: "search err"}
		resp.Res = res
		return resp, nil
	}
	var bss []*filscanproto.Blocksize
	var totalBlockCount int
	var totalBlocksize int64
	var min float64
	var max float64
	for key := range tArr {
		blocksize := new(filscanproto.Blocksize)
		start := tArr[key]
		var end int64
		if key == len(tArr)-1 {
			end = models.TimeNow
		} else {
			end = tArr[key+1]
		}
		bms, err := GetBlockNumByTime(start, end)
		if err != nil {
			log("search err=%v ", err)
			res := &common.Result{Code: 5, Msg: "search err"}
			resp.Res = res
			return resp, nil
		}
		blockCount := len(bms)
		var blockSize int64
		for _, value := range bms {
			blockSize += value.Size
		}

		totalBlockCount += blockCount
		totalBlocksize += blockSize

		blocksize.Time = start
		if blockCount > 0 {
			blocksize.BlockSize = float64(blockSize / int64(blockCount))
		} else {
			blocksize.BlockSize = 0
		}
		if key == 0 {
			max = blocksize.BlockSize
			min = blocksize.BlockSize
		} else {
			if blocksize.BlockSize > max {
				max = blocksize.BlockSize
			}
			if blocksize.BlockSize < min {
				min = blocksize.BlockSize
			}
		}
		bss = append(bss, blocksize)
	}
	avgBlocksize := 0.0
	if totalBlockCount > 0 {
		avgBlocksize = float64(totalBlocksize / int64(totalBlockCount))
	}
	avgBlockSizeCash.Min = min
	avgBlockSizeCash.Max = max
	avgBlockSizeCash.AvgBlocksize = avgBlocksize
	avgBlockSizeCash.Time = endTime
	avgBlockSizeCash.BlockSize = bss
	resp.Data = &filscanproto.AvgBlockheaderSizeGraphicalResp_Data{Data: bss, AvgBlocksize: avgBlocksize, Min: min, Max: max}
	resp.Res = common.NewResult(3, "success")
	return resp, nil
}

func GetIntHour(startTime, endTime int64) (timeArr []int64, err error) {
	if endTime <= startTime {
		return nil, errors.New("endTime <= startTime ")
	}
	var i int64 = 0
	for ; i < (endTime - startTime); i++ {
		if (startTime+i)%3600 == 0 {
			timeArr = append(timeArr, startTime+i)
			break
		}
	}
	if len(timeArr) < 1 {
		return nil, errors.New("Time No IntHour")
	}

	for j := 1; startTime+int64(3600*j) < endTime; j++ {
		timeArr = append(timeArr, timeArr[0]+int64(3600*j))
	}

	return
}

func (this *FilscanServer) FILOutsanding(ctx context.Context, input *common.Empty) (*filscanproto.FILOutsandingResp, error) {
	resp := new(filscanproto.FILOutsandingResp)
	return resp, nil
}

func (this *FilscanServer) LatestBlock(ctx context.Context, input *filscanproto.LatestBlockReq) (*filscanproto.LatestBlockResp, error) {
	resp := new(filscanproto.LatestBlockResp)
	num := input.GetNum()
	if !CheckArg(num) {
		res := &common.Result{Code: 5, Msg: "Missing required parameters"}
		resp.Res = res
		return resp, nil
	}
	v, err := GetBlockByIndex(0, int(num))
	if err != nil {
		resp.Res = &common.Result{Code: 5, Msg: "search err "}
		return resp, nil
	}
	var blockheaders []*filscanproto.FilscanBlock
	for _, value := range v {
		b := FilscanBlockResult2PtotoFilscanBlock(*value)
		blockheaders = append(blockheaders, b)
	}
	resp.Data = &filscanproto.LatestBlockResp_Data{BlockHeader: blockheaders}
	return resp, nil
}

func FilscanBlockResult2PtotoFilscanBlock(f models.FilscanBlockResult) *filscanproto.FilscanBlock {
	res := new(filscanproto.FilscanBlock)
	b := new(filscanproto.BlockHeader)
	b.Miner = f.BlockHeader.Miner
	var parents []string
	//for _, value := range f.BlockHeader.Tickets {
	//	tickets = append(tickets, value.VRFProof)
	//}
	//tickets =

	b.Tickets = f.BlockHeader.Ticket.VRFProof
	b.ElectionProof = f.BlockHeader.ElectionProof
	for _, value := range f.BlockHeader.Parents {
		parents = append(parents, value.Str)
	}
	b.Parents = parents
	b.ParentWeight = f.BlockHeader.ParentWeight
	b.Height = int64(f.BlockHeader.Height)
	b.ParentStateRoot = f.BlockHeader.ParentMessageReceipts.Str
	b.ParentMessageReceipts = f.BlockHeader.ParentMessageReceipts.Str
	b.Messages = f.BlockHeader.Messages.Str
	b.BlsAggregate = &filscanproto.Signature{Type: f.BlockHeader.BLSAggregate.Type, Data: f.BlockHeader.BLSAggregate.Data}
	b.Timestamp = int64(f.BlockHeader.Timestamp)
	b.BlockSig = &filscanproto.Signature{Type: f.BlockHeader.BlockSig.Data, Data: f.BlockHeader.BlockSig.Type}
	res.BlockHeader = b
	res.Cid = f.Cid
	//res.Weight = f.
	res.Size = f.Size
	for _, value := range f.MsgCids {
		res.MsgCids = append(res.MsgCids, value.Str)
	}
	res.Reward = f.BlockReword
	return res
}

func FilscanResMsg2PtotoFilscanMessage(m models.FilscanMsgResult) *filscanproto.FilscanMessage {
	res := new(filscanproto.FilscanMessage)
	res.Size = m.Size
	res.Cid = m.Cid
	res.Msgcreate = int64(m.MsgCreate)
	res.Height = m.Height
	res.BlockCid = m.BlockCid
	res.ExitCode = m.ExitCode
	//res.GasUsed = m.GasUsed
	res.Return = m.Return
	res.MethodName = m.MethodName

	msg := new(filscanproto.Message)
	msg.To = m.Message.To
	msg.From = m.Message.From
	msg.Nonce = m.Message.Nonce
	//f64 ,_ := strconv.ParseFloat(m.Message.Value,64)
	msg.Value = m.Message.Value
	msg.Gasprice = m.Message.GasPrice
	msg.Gaslimit = m.Message.GasLimit
	msg.Method = strconv.Itoa(m.Message.Method)
	msg.Params = m.Message.Params

	res.Msg = msg
	return res
}

func FilscanBlock2FilscanBlockResult(fb []*models.FilscanBlock) (fbr []*models.FilscanBlockResult) {
	for _, value := range fb {
		tbyte, _ := json.Marshal(value)
		var p *models.FilscanBlockResult
		json.Unmarshal(tbyte, &p)
		fbr = append(fbr, p)
	}
	return
}

func FilscanMsg2FilscanMsgResult(fm []*models.FilscanMsg) (fmr []*models.FilscanMsgResult) {
	for _, value := range fm {
		tbyte, _ := json.Marshal(value)
		var p *models.FilscanMsgResult
		json.Unmarshal(tbyte, &p)
		fmr = append(fmr, p)
	}
	return
}

func (this *FilscanServer) LatestMsg(ctx context.Context, input *filscanproto.LatestMsgReq) (*filscanproto.LatestMsgResp, error) {
	resp := new(filscanproto.LatestMsgResp)

	num := input.GetNum()
	if !CheckArg(num) {
		res := &common.Result{
			Code: 5,
			Msg:  "Missing required parameters",
		}
		resp.Res = res
		return resp, nil
	}
	v, err := GetMsgByIndex(0, int(num))
	if err != nil {
		resp.Res = common.NewResult(5, "search err")
		return resp, nil
	}
	var msgheaders []*filscanproto.FilscanMessage
	for _, value := range v {
		b := FilscanResMsg2PtotoFilscanMessage(*value)
		msgheaders = append(msgheaders, b)
	}
	resp.Data = &filscanproto.LatestMsgResp_Data{Msg: msgheaders}
	return resp, nil
}