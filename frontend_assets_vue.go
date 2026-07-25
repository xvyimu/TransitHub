//go:build !frontend_external && frontend_vue

package main

import "embed"

// Vue web-console 资源仅在 `-tags frontend_vue` 构建时嵌入。web-console/dist 由
// 前端构建（pnpm --dir web-console build）产出，不入库；因此默认构建不引用它，
// 只有显式请求 Vue 交付面的构建才要求该目录在编译时存在。
//
//go:embed web-console/dist
var vueBuildFS embed.FS

//go:embed web-console/dist/index.html
var vueIndexPage []byte

// vueThemeAssets 返回嵌入的 Vue web-console 资源，供 FRONTEND_MODE=vue 使用。
func vueThemeAssets() (embed.FS, []byte) {
	return vueBuildFS, vueIndexPage
}
