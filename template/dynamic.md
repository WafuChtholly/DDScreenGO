# B站动态模板 (dynamic1) 变量字典

在 `dynamic1` 变体中（对应的模板文件应放在 `template/bilidynamic1.tmpl` 中），系统通过 B 站 API 获取了该动态的**最原始 JSON 数据**，并通过 `{{.RawData}}` 变量直接暴露给你。

你可以用你在浏览器或 curl 获取到的 B 站原始 JSON 结构，通过 `{{.RawData.data.item.XXX}}` 的语法随意读取任何一层的数据！

由于 B 站接口返回的内容极多，这里只列出最核心的底层变量路径速查字典：

## 模板内置控制变量 (非 RawData)
除了 `{{.RawData}}`，系统还为模板注入了以下几个顶级控制变量，用于实现特定功能：
- `{{.Expand}}` (bool): 用户是否开启了“强制展开所有图片”。开启时，你应当将图片渲染为垂直列表而不是九宫格。
- `{{.AtCard}}` (bool): 用户是否开启了“展开 @ 用户卡片”。
- `{{.LinkQr}}` (bool): 用户是否开启了“展开网页链接二维码”。

## 基础结构
动态的所有核心内容都包裹在 `{{.RawData.data.item.modules}}` 中。它通常包含了几个核心模块：
- `module_author`：作者信息（头像、名字、粉丝牌、发布时间等）
- `module_dynamic`：动态内容（正文文本、富文本、图片、投票、话题等）

---

## 1. 作者信息 `module_author`
- `{{.RawData.data.item.modules.module_author.name}}` - 作者的用户名（如：哈鹿hallu）
- `{{.RawData.data.item.modules.module_author.face}}` - 作者的头像图片链接
- `{{.RawData.data.item.modules.module_author.pub_time}}` - 动态发布时间文本（如："2026年04月11日 19:51"）

**粉丝牌与装扮 (`decorate`)**
部分 UP 主有高贵的粉丝编号或装扮卡片，它们位于 `module_author.decorate` 中：
- `{{.RawData.data.item.modules.module_author.decorate.name}}` - 装扮名字
- `{{.RawData.data.item.modules.module_author.decorate.card_url}}` - 头像右侧长条装饰卡片的背景图
- `{{.RawData.data.item.modules.module_author.decorate.fan.num_str}}` - 高贵数字编号（如图中展示的 "000002"）
- `{{.RawData.data.item.modules.module_author.decorate.fan.color}}` - 编号的颜色代码（如 "#E5B261"）

---

## 2. 动态主内容 `module_dynamic`

**话题 (`topic`)**
- `{{.RawData.data.item.modules.module_dynamic.topic.name}}` - 动态关联的话题文本。

**文字正文 (`desc`)**
- `{{.RawData.data.item.modules.module_dynamic.desc.text}}` - 纯文本的正文内容。

如果需要处理蓝色的 **@提到的人**、**#话题#** 或是**官方表情包**，你需要遍历 `rich_text_nodes` 数组：
```html
{{range $node := .RawData.data.item.modules.module_dynamic.desc.rich_text_nodes}}
    {{if eq $node.type "RICH_TEXT_NODE_TYPE_TEXT"}}
       <span>{{$node.text}}</span>
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_AT"}}
       <!-- 渲染蓝色的 @ 元素 -->
       <span class="bili-link">@{{$node.text}}</span>
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_TOPIC"}}
       <!-- 渲染蓝色的 话题 元素 -->
       <span class="bili-link">#{{$node.text}}#</span>
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_EMOJI"}}
       <!-- 将表情包文字替换为图片 -->
       <img src="{{$node.emoji.icon_url}}" />
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_VOTE"}}
       <!-- 投票标记 -->
       <span class="bili-link">{{$node.text}}</span>
    {{end}}
{{end}}
```

**网页链接 (`RICH_TEXT_NODE_TYPE_WEB`)**
对于网页链接，若想兼容后台自动注入二维码卡片（当开启 LinkQr 时），请务必保证你输出了带有 `bili-link`、`opus-text-rich-hl link` 和正确 href 属性的 `<a>` 标签。
```html
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_WEB"}}
       <!-- 渲染网页链接，附带正确类名以支持二维码注入 -->
       <a class="opus-text-rich-hl link bili-link" href="{{if $node.jump_url}}{{$node.jump_url}}{{else}}#{{end}}" target="_blank">
           网页链接
       </a>
```

**@用户 (`RICH_TEXT_NODE_TYPE_AT`)**
若想兼容后台自动获取该用户信息的卡片（当开启 AtCard 时），请务必给 @ 标签增加 `bili-rich-text-module at` 类，并附加 `data-oid="{{$node.rid}}"`：
```html
    {{else if eq $node.type "RICH_TEXT_NODE_TYPE_AT"}}
       <!-- 渲染蓝色的 @ 元素，附带正确类名与 rid 以支持用户信息卡片注入 -->
       <span class="bili-rich-text-module at bili-link" data-oid="{{$node.rid}}">@{{$node.text}}</span>
```

**图片媒体 (`major.draw`)**
如果这是一条带图动态，图片会存在 `major.draw.items` 数组里。为了响应强制展开开关，推荐如下写法：
```html
{{if .Expand}}
    <!-- 如果启用了强制展开，显示竖排大图 -->
    <div class="image-grid expand-img">
        {{range $pic := .RawData.data.item.modules.module_dynamic.major.draw.items}}
            <img src="{{$pic.src}}" />
        {{end}}
    </div>
{{else}}
    <!-- 否则显示九宫格裁切小图 -->
    <div class="image-grid">
        {{range $index, $pic := .RawData.data.item.modules.module_dynamic.major.draw.items}}
            {{if lt $index 9}}
            <img src="{{$pic.src}}@400w_400h_1c.webp" />
            {{end}}
        {{end}}
    </div>
{{end}}
```

**投票卡片 (`additional.vote`)**
如果动态有发起投票：
- `{{.RawData.data.item.modules.module_dynamic.additional.vote.desc}}` - 投票标题
- `{{.RawData.data.item.modules.module_dynamic.additional.vote.join_num}}` - 参与投票人数

遍历投票选项 `options`：
```html
{{range $opt := .RawData.data.item.modules.module_dynamic.additional.vote.options}}
    <div>选项: {{$opt.opt_desc}}</div>
    <div>票数: {{$opt.cnt}}</div>
{{end}}
```

**直播预约卡片 (`additional.reserve`)**
如果这是一条直播预约动态（且没有写任何正文文本，那么 `desc` 会是 null！）：
- `{{.RawData.data.item.modules.module_dynamic.additional.reserve.title}}` - 预约标题
- `{{.RawData.data.item.modules.module_dynamic.additional.reserve.desc1.text}}` - 直播时间描述（如 "明天 19:00 直播"）
- `{{.RawData.data.item.modules.module_dynamic.additional.reserve.desc2.text}}` - 预约人数描述（如 "579人预约"）

---

## 3. 转发动态处理 (`type: DYNAMIC_TYPE_FORWARD`)
如果最外层的 `{{.RawData.data.item.type}}` 等于 `"DYNAMIC_TYPE_FORWARD"`，则表示这是一条转发动态！

此时：
* 外层的 `module_dynamic.desc.text` 代表转发者的文字（如“芜湖又可以一起玩啦”）。
* **被转发的原动态的全部信息**，会存在一个平级的字段 `orig` 中！

您可以直接用完全相同的嵌套路径去读取 `orig` 里的信息来画出“内嵌引用框”：
- 原作者：`{{.RawData.data.item.orig.modules.module_author.name}}`
- 原动态富文本：`{{.RawData.data.item.orig.modules.module_dynamic.desc.rich_text_nodes}}`
- 原动态图片：`{{.RawData.data.item.orig.modules.module_dynamic.major.draw.items}}`
- 原动态投票：`{{.RawData.data.item.orig.modules.module_dynamic.additional.vote}}`
