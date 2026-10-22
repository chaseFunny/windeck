<p align="center">
  <img src="assets/icon.png" width="128" height="128" alt="窗签图标">
</p>

<h1 align="center">窗签 · Windeck</h1>

<p align="center">
  <b>像切浏览器标签页一样切换 Mac 窗口。</b><br>
  一款 macOS 窗口切换器：按住 <kbd>⌥</kbd> 再按 <kbd>Tab</kbd>，所有窗口按应用分组、以完整标题的标签呈现，中央大预览帮你确认“就是它”。
</p>

<p align="center">
  <a href="https://github.com/chaseFunny/windeck/releases/latest/download/Windeck-arm64.dmg"><b>下载 macOS 版</b></a> ·
  <a href="https://windeck.islumi.com/zh">官网</a> ·
  <a href="README.md">English</a>
</p>

<p align="center">
  <img src="assets/switcher.png" alt="窗签切换面板：窗口按应用分组列为标签，中央显示选中窗口的大预览" width="860">
</p>

## 为什么需要窗签

macOS 自带的 <kbd>⌘</kbd> <kbd>Tab</kbd> 切换的是**应用**，不是窗口。同时开着三个浏览器窗口、两个编辑器和一个终端时，想找到某一个窗口只能靠猜。窗签是 **Mac 上的 Alt-Tab**，按**窗口**切换，就像在浏览器里切标签页：

- **完整标题，按应用分组**：左侧纵向标签栏让每个窗口标题完整可读，多窗口的应用折叠为分组。
- **中央大预览**：选中的窗口大图显示，两个相似的窗口再也不会认错。
- **不管窗口在哪**：最小化、已隐藏、全屏、在其他桌面（Space）或外接显示器的窗口全部列出，并用徽标标明位置。
- **位置稳定不乱跳**：应用与窗口在列表中的位置固定，只有高亮在移动。
- **不离开面板就能处理**：<kbd>W</kbd> 关闭窗口、<kbd>M</kbd> 最小化、<kbd>Q</kbd> 退出应用。
- **搜索**：按 <kbd>空格</kbd>，输入标题里的几个字。

## 怎么用

| 按键 | 效果 |
|---|---|
| 快速按一下 <kbd>⌥</kbd> <kbd>Tab</kbd> | 回到上一个窗口，屏幕上什么都不会弹出 |
| 按住 <kbd>⌥</kbd>，按 <kbd>Tab</kbd> | 标签栏出现；<kbd>Tab</kbd> 按最近使用顺序跳转，方向键按列表移动；松开即切换 |
| <kbd>⌥</kbd> <kbd>Tab</kbd> 后按 <kbd>空格</kbd> | 搜索：输入后按 <kbd>回车</kbd> |

推荐保留系统 <kbd>⌘</kbd> <kbd>Tab</kbd> 切换应用、用 <kbd>⌥</kbd> <kbd>Tab</kbd> 切换窗口；也可以让窗签全面接管 <kbd>⌘</kbd> <kbd>Tab</kbd>。全部按键见 [docs/shortcuts.md](docs/shortcuts.md)。

## 智能，但从不碍事

- 第 1 顺位**永远**是上一个窗口，盲按一下 <kbd>Tab</kbd> 一定回到刚才的窗口。
- 第 2 顺位是你通常接下来会去的窗口，带一个小星标。
- 搜索结果综合使用频率和最近使用时间排序。

只在本机学习。详见 [docs/smart.md](docs/smart.md)。

## 隐私

- 窗口标题、画面、文件名、应用列表**绝不离开你的 Mac**。
- 隐私排除名单：默认预置密码管理器与网银客户端，从不截取。
- 一键清除全部学习数据与快照。
- 本地数据格式公开：[docs/data.md](docs/data.md)。

## 安装

1. [下载 Windeck-arm64.dmg](https://github.com/chaseFunny/windeck/releases/latest/download/Windeck-arm64.dmg)，把 **Windeck** 拖进“应用程序”。
2. 打开它。首个版本还没有经过 Apple 公证，macOS 可能提示无法验证：打开 **系统设置 → 隐私与安全性**，点 **仍要打开**；或者在终端运行
   ```sh
   xattr -dr com.apple.quarantine /Applications/Windeck.app
   ```
3. 按提示允许 **辅助功能**（列出和切换窗口需要）；**屏幕录制** 可选，只用于显示预览。
4. 选择快捷键，开始使用。

系统要求：macOS 13 Ventura 及以上，Apple 芯片。详见 [docs/quick-start.md](docs/quick-start.md)。

## 价格

免费版即可完整切换：标签、预览、快切、搜索、各类窗口。**窗签专业版** 增加第 2 顺位预测与按习惯排序的搜索；每次安装都有 14 天专业版试用。

## 常见问题

**是 Mac 上的 Alt-Tab / AltTab 替代品吗？** 是的：窗签按单个窗口切换，就像 Windows 上的 Alt-Tab，并且按应用分组显示标题和大预览。

**会替换系统的 ⌘Tab 吗？** 由你决定：保留 ⌘Tab 切换应用、用 ⌥Tab 切换窗口，或全面接管 ⌘Tab。

**为什么需要辅助功能权限？** 用于列出窗口、激活窗口，以及感知快捷键的按下与松开。macOS 上所有窗口切换器都需要。

**必须开启屏幕录制吗？** 不必须，拒绝时只显示标题。

**会帮我平铺、排列窗口吗？** 不会，窗签面向 macOS 默认的堆叠式窗口。

**问题反馈和建议？** 请到 [Issues](https://github.com/chaseFunny/windeck/issues)。

---

窗签由 [hz](https://github.com/hz7002) 出品。官网：[windeck.islumi.com](https://windeck.islumi.com/zh)。

<sub>关键词：Mac 窗口切换、macOS 窗口切换器、Mac Alt Tab、AltTab 替代、Cmd Tab 替代、Mac 多窗口切换、Mac 窗口管理工具。</sub>
