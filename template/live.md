# DDScreenGO 模板参数说明文档 (example.md)

这份文档详细说明了在自定义模板（`.tmpl` 文件）中可以使用的所有参数。
无论你从 Swagger API 传入什么样的自定义参数，后端的 Go 程序都会将它们抓取、整理，并打包成一个统一的 `LiveTemplateData` 结构体发送给模板。这样在模板中就可以以最简单的方式调用这些数据。

---

## 1. 原理与实现过程：输入参数 -> 模板变量

当你发起一个请求，比如：
`GET /api/Bili/Live4?roomid=1808576670&live_state=0&qr=true&tips=自定义文本`

1. **接收与解析**：后端接收到请求后，提取出 `roomid`, `live_state`, `qr`, `tips` 等输入参数，并将它们存储在后端的 `CardOptions` 配置对象中。
2. **抓取上游数据**：后端拿着 `roomid` 去请求对应的平台（比如 B站 API），抓取主播的标题、分区、头像、封面、粉丝数等信息，存储在 `platform.LiveInfo` 对象中。
3. **图像处理**：将封面（Cover）和头像（Avatar）等图片下载并转换成 `Base64` 编码（无需依赖网络，直接在 HTML 中显示）；如果需要二维码，也会生成二维码的 `Base64` 字符串。
4. **组装**：把上述所有的对象组合成一个超级字典 `LiveTemplateData`，传给模板引擎。
5. **渲染**：我们在 `.tmpl` 文件中，通过 Go 语言内置的 `{{.变量名}}` 语法即可自动替换对应的数据，生成最终的 HTML 给无头浏览器截图。

---

## 2. 可以在模板中使用的参数全集

以下所有参数都可以在你的 `bililive4.tmpl` (或者其他平台的模板) 中直接通过 `{{.参数名}}` 的形式来调用！

### 🍎 A. 主播基本信息 `{{.Info.xxx}}`
这部分数据是程序去官方接口拉取的，你只需要提供 `roomid` 就能获取。

* **`{{.Info.Title}}`**：直播间标题。
* **`{{.Info.RoomID}}`**：直播间号码 / 用户ID。
* **`{{.Info.Nickname}}`** / **`{{.Info.Author}}`**：主播的昵称或名字。
* **`{{.Info.Category}}`**：直播的分区（如 "娱乐-颜值"）。
* **`{{.Info.Description}}`**：直播间的简介（如果有）。
* **`{{.Info.FollowerNum}}`**：主播的粉丝数或直播间人气值。
* **`{{.Info.Platform}}`**：当前平台名称（如 "Bilibili", "Weibo"）。

### 🍊 B. API 用户自定义参数 `{{.Options.xxx}}`
这部分数据**完全对应你在 API 页面（或 Swagger 图一）中手动传入的参数**。

* **`{{.Options.Tips}}`**：对应 API 传入的 `tips`。可以作为自定义联名文本展示。
* **`{{.Options.ModelOrder}}`**：对应 API 传入的 `model_order`。你可以在模板中用它来控制模块显示的顺序。
* **`{{.Options.View}}`**：对应 API 传入的 `view`。一般 `0` 表示桌面端截图，`1` 表示移动端，你可以在模板里用它做条件判断（改变长宽比等）。
* **`{{.Options.LiveState}}`**：对应 API 传入的 `live_state` (0: 直播中, 1: 下播)。
* **`{{.Options.Timestamp}}`**：对应传入的时间戳。

### 🫐 C. 后端处理好的状态和图片 (顶级变量)
这部分是后端为你“加工好”的高级数据，无需你写 JS 来转换。

* **`{{.CoverBase64}}`**：直播间封面的 Base64 图片数据（可直接用于 `<img src="...">`）。
* **`{{.AvatarBase64}}`**：主播头像的 Base64 图片数据。
* **`{{.QRBase64}}`**：自动生成的二维码 Base64 图片数据。只有在 API 传入 `qr=true` 时才有值，如果传入 false 就是空字符串。
* **`{{.IsLive}}`**：布尔值 (`true` / `false`)。后端根据你传入的 `live_state` 自动算出的。如果为 true，表示正在直播。
* **`{{.Duration}}`**：如果未开播，会显示已经开播或下播多长时间（比如 "2 小时 30 分钟"）。
* **`{{.DurationLabel}}`**："直播时长：" 这样的前缀文本。
* **`{{.GeneratedAt}}`**：该图片生成时的系统当前时间（格式: "2026-06-04 12:30:00"）。

---

## 3. 在模板中的实际使用范例

下面这个简单的 HTML 片段展示了如何在实战中使用上面提到的所有参数。

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>{{.Info.Title}}</title>
</head>
<body>
    <!-- 1. 使用处理好的 Base64 图片 (自动规避防盗链) -->
    <img src="{{.AvatarBase64}}" alt="头像" class="avatar">
    <img src="{{.CoverBase64}}" alt="封面" class="cover">

    <!-- 2. 使用 API 传入的 自定义参数 tips -->
    <div class="custom-tips">这是联名提示语：{{.Options.Tips}}</div>
    
    <!-- 利用 View 参数做多端适配的判断 (支持简单的 if 条件语句) -->
    {{if eq .Options.View 1}}
        <p>当前是移动端截图模式</p>
    {{else}}
        <p>当前是电脑端截图模式</p>
    {{end}}

    <!-- 3. 使用抓取到的 主播信息 -->
    <h1>{{.Info.Nickname}} 的直播间</h1>
    <p>当前分区：{{.Info.Category}}</p>
    <p>粉丝数量：{{.Info.FollowerNum}}</p>

    <!-- 4. 根据后台算好的 IsLive 判断上下播状态 -->
    {{if .IsLive}}
        <div class="badge">🔥 正在直播</div>
    {{else}}
        <div class="badge">💤 休息中</div>
        <p>{{.DurationLabel}} {{.Duration}}</p>
    {{end}}

    <!-- 5. 如果 API 传入 qr=true，QRBase64 就有值，否则为空 -->
    {{if .QRBase64}}
        <div class="qr-code">
            <img src="{{.QRBase64}}" alt="扫描二维码观看">
        </div>
    {{end}}

    <!-- 打印生成时间 -->
    <p>图片生成时间: {{.GeneratedAt}}</p>
</body>
</html>
```

看完这份文档，你应该能完全明白数据是怎么从 **API 输入 -> 后端抓取加工 -> 模板展示** 的完整链路了！
