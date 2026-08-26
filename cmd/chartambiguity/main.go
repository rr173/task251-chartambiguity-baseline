// 命令入口：学术图表视觉编码歧义复核台。
// 提供两种运行模式：
//   - 服务模式：--addr :端口 --db 路径，启动 /api HTTP 服务（含浏览器复核页）。
//   - 自检模式：--smoke-test，真实创建图形→导入语义→声明编码→检测歧义→
//     豁免→发布规范→关闭并重开数据库验证持久化与重启恢复，最后以 0 退出。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"task251-chartambiguity/internal/httpapi"
	"task251-chartambiguity/internal/service"
	"task251-chartambiguity/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "chartambiguity.db", "SQLite 数据库路径")
	smoke := flag.Bool("smoke-test", false, "运行自检后退出，验证持久化与重启恢复")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(); err != nil {
			log.Fatalf("smoke-test FAILED: %v", err)
		}
		fmt.Println("smoke-test PASSED")
		return
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	api := httpapi.New(svc)
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 极简浏览器复核页：列出图形稿与自检入口。
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, pageHTML)
	})

	server := &http.Server{Addr: *addr, Handler: mux}
	fmt.Printf("task251-chartambiguity listening on %s (db=%s)\n", *addr, *dbPath)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// runSmokeTest 验证完整业务闭环与重启恢复。
func runSmokeTest() error {
	dir, err := os.MkdirTemp("", "chartambiguity-smoke")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbFile := dir + "/smoke.db"

	// ---- 第一次会话：创建→导入→声明→检测→豁免→发布 ----
	st, err := store.Open(dbFile)
	if err != nil {
		return err
	}
	svc := service.New(st)

	fig, err := svc.CreateFigure("烟雾测试图：温度与浓度共用蓝色")
	if err != nil {
		return fmt.Errorf("create figure: %w", err)
	}
	fid := fig.ID

	payload := service.ImportPayload{
		Layers:    []service.LayerInput{{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Axes:      []service.AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: "x"}, {Name: "y", Variable: "temp", Unit: "K", Orientation: "y"}},
		Variables: []service.VariableInput{{Name: "temp", Unit: "K"}, {Name: "conc", Unit: "mol/L"}},
		Legends:   []service.LegendInput{{Channel: "color", Label: "temperature", Token: "#1f77b4", CoversVariable: "temp"}},
	}
	if err := svc.ImportSemantics(fid, payload); err != nil {
		return fmt.Errorf("import: %w", err)
	}
	// 温度与浓度均用同一蓝色 #1f77b4 → 颜色复用歧义。
	if _, err := svc.DeclareEncoding(fid, "", "temp", "color", "#1f77b4"); err != nil {
		return fmt.Errorf("encode temp: %w", err)
	}
	if _, err := svc.DeclareEncoding(fid, "", "conc", "color", "#1f77b4"); err != nil {
		return fmt.Errorf("encode conc: %w", err)
	}
	ambigs, err := svc.RunCheck(fid)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	open := 0
	for _, a := range ambigs {
		if a.IsOpen() {
			open++
		}
	}
	if open != 1 {
		return fmt.Errorf("expected 1 open ambiguity, got %d", open)
	}
	fmt.Printf("  [smoke] detected %d open ambiguity(ies) as expected\n", open)

	// 豁免该颜色复用，重算后应全部解决。
	if _, err := svc.AddException(fid, service.ExceptionInput{
		Kind:          "reuse_exemption",
		TargetChannel: "color",
		TargetToken:   "#1f77b4",
		Reason:        "作者确认温度与浓度同色表示不同量纲",
	}); err != nil {
		return fmt.Errorf("exception: %w", err)
	}
	ambigs, err = svc.RunCheck(fid)
	if err != nil {
		return fmt.Errorf("recheck: %w", err)
	}
	for _, a := range ambigs {
		if a.IsOpen() {
			return fmt.Errorf("ambiguity still open after exemption: %s", a.Description)
		}
	}
	fmt.Println("  [smoke] ambiguity resolved by exception")

	// 发布规范版本并冻结图形稿。
	spec1, err := svc.CreateSpec(fid)
	if err != nil {
		return fmt.Errorf("create spec: %w", err)
	}
	if _, err := svc.PublishSpec(fid, spec1.ID); err != nil {
		return fmt.Errorf("publish spec: %w", err)
	}
	f, err := svc.GetFigure(fid)
	if err != nil {
		return err
	}
	if f.Status != "frozen" {
		return fmt.Errorf("expected frozen, got %s", f.Status)
	}
	fmt.Println("  [smoke] spec published and figure frozen")

	if err := st.Close(); err != nil {
		return err
	}

	// ---- 第二次会话：重开同一数据库，验证持久化与重启恢复 ----
	st2, err := store.Open(dbFile)
	if err != nil {
		return fmt.Errorf("reopen db: %w", err)
	}
	defer st2.Close()
	svc2 := service.New(st2)

	f2, err := svc2.GetFigure(fid)
	if err != nil {
		return fmt.Errorf("reopen get figure: %w", err)
	}
	if f2.Status != "frozen" {
		return fmt.Errorf("after restart expected frozen, got %s", f2.Status)
	}
	encs, err := st2.ListEncodings(fid)
	if err != nil {
		return err
	}
	if len(encs) != 2 {
		return fmt.Errorf("after restart expected 2 encodings, got %d", len(encs))
	}
	specs, err := st2.ListSpecs(fid)
	if err != nil {
		return err
	}
	if len(specs) != 1 || specs[0].Status != "frozen" {
		return fmt.Errorf("after restart spec not frozen: %+v", specs)
	}
	fmt.Println("  [smoke] restart recovery verified (figure/spec/encodings persisted)")
	return nil
}

// pageHTML 是极简浏览器复核页（非前端项目，仅便捷入口）。
const pageHTML = `<!doctype html><html lang="zh"><head><meta charset="utf-8">
<title>学术图表视觉编码歧义复核台</title></head><body>
<h1>学术图表视觉编码歧义复核台</h1>
<p>服务已启动。可用端点：</p>
<ul>
<li><code>POST /api/figures</code> 创建图形稿</li>
<li><code>POST /api/figures/{id}/import</code> 导入语义</li>
<li><code>POST /api/figures/{id}/encodings</code> 声明编码</li>
<li><code>POST /api/figures/{id}/check</code> 复核歧义</li>
<li><code>GET /api/figures/{id}/ambiguities</code> 查看歧义</li>
<li><code>POST /api/figures/{id}/exceptions</code> 登记豁免</li>
<li><code>POST /api/figures/{id}/specs</code> 创建规范</li>
<li><code>POST /api/figures/{id}/specs/{sid}/publish</code> 冻结发布</li>
<li><code>GET /api/selfcheck</code> 服务自检</li>
</ul>
</body></html>`
