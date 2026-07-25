//go:build !frontend_external && !frontend_vue

package main

import "embed"

// vueThemeAssets 的默认（无 frontend_vue tag）实现：不嵌入 web-console/dist，
// 返回空资源。这样干净 checkout 上 `go build ./...` 不依赖前端构建产物；
// 若运行时设 FRONTEND_MODE=vue，registerVueFrontend 会因 VueIndexPage 为空而
// 明确报错，而不是编译失败。要真正嵌入 Vue 控制台，用 `-tags frontend_vue` 构建。
func vueThemeAssets() (embed.FS, []byte) {
	return embed.FS{}, nil
}
