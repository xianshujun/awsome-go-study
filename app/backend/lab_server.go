package backend

import (
	"net/http"
	"strings"
	"time"
)

/*
/api/lab 是事故演练场的唯一入口，action 决定做什么：

	bootstrap  今日事故、历史存档、场景库与本机成绩
	case       打开某一天/某个场景，返回服务、指标、告警与已揭示的提示
	run        在模拟终端里执行一条只读排查命令
	submit     提交根因与处置方案，返回评分与复盘
	hint       揭示下一条提示（会扣分）
	reset      清空本次演练的终端记录与提示，保留历史最好成绩

所有命令都在内存里的演练世界里执行，不会碰真实文件、网络或进程。
*/
func (e *Engine) handleLab(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action    string `json:"action"`
		Scenario  string `json:"scenario"`
		Date      string `json:"date"`
		CaseID    string `json:"caseId"`
		Command   string `json:"command"`
		RootCause *int   `json:"rootCause"`
		Fix       *int   `json:"fix"`
	}
	if !decode(w, r, &req) {
		return
	}
	switch strings.TrimSpace(req.Action) {
	case "bootstrap", "":
		s := e.store.snapshot()
		modelReady := strings.TrimSpace(s.Settings.BaseURL) != "" && strings.TrimSpace(s.Settings.Model) != ""
		writeJSON(w, http.StatusOK, e.labs.bootstrap(s.Lab, modelReady, time.Now()))
	case "case":
		s := e.store.snapshot()
		view, err := e.labs.openCase(req.Scenario, req.Date, s.Lab, time.Now())
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, view)
	case "run":
		result, err := e.labs.run(req.CaseID, req.Command)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "submit":
		if req.RootCause == nil || req.Fix == nil {
			writeError(w, http.StatusBadRequest, "请选择根因与处置方案")
			return
		}
		s := e.store.snapshot()
		result, updated, err := e.labs.submit(req.CaseID, *req.RootCause, *req.Fix, s.Lab)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if _, err := e.store.update(func(st *state) error {
			st.Lab = updated
			return nil
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "保存演练成绩失败")
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "hint":
		result, err := e.labs.hint(req.CaseID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "reset":
		if err := e.labs.reset(req.CaseID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
	default:
		writeError(w, http.StatusBadRequest, "未知的演练操作")
	}
}
