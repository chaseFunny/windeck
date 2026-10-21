// The site's words, in English and Simplified Chinese: one page, two
// languages, the same layout.

export type Lang = 'en' | 'zh';

export const DOWNLOAD_URL = 'https://github.com/windeck/windeck/releases/latest';
export const BUY_URL = 'https://windeck.lemonsqueezy.com/checkout';

export const content = {
  en: {
    nav: { features: 'Features', smart: 'Smart', privacy: 'Privacy', pricing: 'Pricing', faq: 'FAQ', download: 'Download', other: '中文', otherHref: '/zh' },
    hero: {
      eyebrow: 'A window switcher for macOS',
      title: ['Your windows,', 'like browser tabs.'],
      lead: 'Hold ⌥ and press Tab. Every window shows as a readable tab, grouped by app, with a large preview to be sure. Let go, and you are there.',
      download: 'Download for macOS',
      brew: 'brew install --cask windeck',
      note: 'Free · macOS 13 or later · Apple silicon and Intel',
    },
    steps: [
      { keys: ['⌥', '⇥'], title: 'Tap to go back', body: 'A quick press and release returns to the previous window. Nothing flashes on screen.' },
      { keys: ['⌥', '⇥', '⇥'], title: 'Hold to look', body: 'Keep holding and the deck appears. Tab moves along your recent windows, arrows along the list.' },
      { keys: ['⌥', '␣'], title: 'Type to find', body: 'Press Space and let go of ⌥. Type a few letters of any title; Return switches.' },
    ],
    featuresTitle: 'Find it fast. Be sure it is the one.',
    featuresLead: 'Tabs do the finding, the preview does the confirming.',
    features: [
      { title: 'Full titles, grouped by app', body: 'A vertical tab bar keeps every title readable. Apps with several windows fold into a group; the rest stay single tabs.' },
      { title: 'A large, sharp preview', body: 'The highlighted window shows big in the middle, so two similar editor windows never get mixed up.' },
      { title: 'Every window, wherever it is', body: 'Minimized, hidden, on another desktop or another display: all listed, each with a badge saying where.' },
      { title: 'Places that stay put', body: 'Apps and windows keep their place in the list; only the highlight moves. Your spatial memory holds.' },
      { title: 'Act without leaving', body: 'W closes, M minimizes, Q quits the app, while the panel stays and the highlight moves on.' },
      { title: 'Light, fast, native', body: 'A single native binary of a few megabytes. Event-driven, no polling: it idles at 0% CPU.' },
    ],
    smart: {
      title: 'Smart where it helps. Never in the way.',
      lead: 'Windeck learns which window you go to next, on your Mac only, and never touches what your fingers rely on.',
      points: [
        { title: 'The first Tab is always the previous window', body: 'No prediction ever takes the first place. A blind tap goes back, every time.' },
        { title: 'The second place is predicted', body: 'From the windows you usually go to next, the likely one waits one Tab further, marked with a small star.' },
        { title: 'Search ranked by habit', body: 'The same letters bring your usual window first, by how often and how lately you use it.' },
      ],
      cycle: ['Previous window', 'Predicted', 'Recent', 'Recent'],
    },
    privacy: {
      title: 'Your windows stay yours.',
      lead: 'Pictures, habits and what Windeck learns are kept on this Mac, in files you can read.',
      points: [
        'No window titles, pictures, file names or app lists ever leave your Mac',
        'A privacy list: password managers and banking apps are never captured, by default',
        'One click forgets everything learned and every picture kept',
        'Without Screen Recording, Windeck works as a plain list of titles',
      ],
    },
    perf: {
      title: 'Speed is the floor.',
      items: [
        { value: '<16 ms', label: 'a blind quick switch' },
        { value: '<50 ms', label: 'from shortcut to panel' },
        { value: '0%', label: 'CPU while idle' },
        { value: '<150 MB', label: 'memory, pictures included' },
      ],
    },
    pricing: {
      title: 'Free to switch. Pro to go further.',
      lead: 'Every install starts with a 14-day Pro trial. No card needed.',
      free: { name: 'Free', price: '$0', note: 'forever', items: ['Grouped tabs and large previews', 'Quick switch and keyword search', 'Minimized, hidden and other-desktop windows', 'Close, minimize and quit from the panel'] },
      pro: { name: 'Pro', price: '$12.99', note: 'once · ¥68 in China', items: ['Everything in Free', 'Predicted second window', 'Search ranked by your habits', 'Coming: pinyin search, layouts, pinned apps, multi-display filters', 'One license, three Macs'] },
      buy: 'Buy Pro',
      early: 'Early-bird price $9.99 / ¥48 at launch. 14-day refund.',
    },
    faq: [
      { q: 'Why does Windeck need Accessibility?', a: 'To list your windows, bring one forward, and hear the shortcut being held and released. macOS asks for it for every window switcher.' },
      { q: 'Is Screen Recording required?', a: 'No. It is only used to show previews. Without it, Windeck shows titles alone, which still works well.' },
      { q: 'Does it replace ⌘Tab?', a: 'Your choice. The recommended mode keeps ⌘Tab for apps and uses ⌥Tab for windows; the takeover mode uses ⌘Tab for windows.' },
      { q: 'What about tiling window managers?', a: 'Windeck is for the usual stacked windows of macOS. It neither tiles nor arranges windows.' },
      { q: 'Is it on the Mac App Store?', a: 'No: the permissions a window switcher needs are not allowed there. Download it here or with Homebrew; it updates itself.' },
    ],
    cta: { title: 'Try it for a day. Keep it for years.', button: 'Download Windeck' },
    footer: { tagline: 'Windeck · 窗签 — your windows, at hand like a deck of cards.', privacy: 'Privacy', contact: 'Contact' },
    demo: { search: 'Search windows', space: 'Space', hints: ['Next', 'Move', 'Close', 'Search'], minimized: 'Minimized', desktop: 'Desktop 2', predicted: 'Predicted' },
  },
  zh: {
    nav: { features: '功能', smart: '智能', privacy: '隐私', pricing: '价格', faq: '常见问题', download: '下载', other: 'English', otherHref: '/' },
    hero: {
      eyebrow: 'macOS 窗口切换器',
      title: ['你的窗口，', '像浏览器标签页一样切换。'],
      lead: '按住 ⌥ 再按 Tab，所有窗口按应用分组、以完整标题的标签呈现，中央大预览帮你确认“就是它”。松开即到达。',
      download: '下载 macOS 版',
      brew: 'brew install --cask windeck',
      note: '免费 · macOS 13 及以上 · 支持 Apple 芯片与 Intel',
    },
    steps: [
      { keys: ['⌥', '⇥'], title: '轻按，回到上一个窗口', body: '快速按下并松开，直接回到刚才的窗口，屏幕上什么都不会闪。' },
      { keys: ['⌥', '⇥', '⇥'], title: '按住，看清所有窗口', body: '继续按住，窗签展开。Tab 按最近使用顺序跳转，方向键按列表上下移动。' },
      { keys: ['⌥', '␣'], title: '打字，找到它', body: '按空格后即可松开 ⌥，输入标题里的几个字，回车即切换。' },
    ],
    featuresTitle: '找得快，也认得准。',
    featuresLead: '标签负责找，预览负责确认。',
    features: [
      { title: '完整标题，按应用分组', body: '左侧纵向标签栏让每个标题完整可读。多窗口的应用折叠为分组，单窗口的应用直接显示为一行。' },
      { title: '高清大预览', body: '选中的窗口在中央大图显示，两个相似的编辑器窗口再也不会认错。' },
      { title: '不管窗口在哪', body: '最小化、已隐藏、在其他桌面或外接显示器的窗口全部列出，并用徽标标明位置。' },
      { title: '位置稳定不乱跳', body: '应用与窗口在列表中的位置固定，只有高亮在移动，保护你的空间记忆。' },
      { title: '不离开面板就能处理', body: 'W 关闭、M 最小化、Q 退出应用，面板保持不动，高亮自动顺移下一项。' },
      { title: '轻巧、迅速、原生', body: '只有几 MB 的原生单文件应用，纯事件驱动、不轮询，空闲时 CPU 为 0%。' },
    ],
    smart: {
      title: '智能只在需要时出现，从不碍事。',
      lead: '窗签在本机学习你接下来常去的窗口，但绝不改动你的肌肉记忆。',
      points: [
        { title: '第 1 顺位永远是上一个窗口', body: '任何预测都不会占用第 1 顺位，盲按一下 Tab 永远回到刚才的窗口。' },
        { title: '第 2 顺位智能预测', body: '根据你的切换习惯，最可能要去的窗口就在下一个 Tab，并带有小星标提示。' },
        { title: '搜索按习惯排序', body: '输入同样的字母，综合使用频率与最近使用时间，优先显示你最常去的那个窗口。' },
      ],
      cycle: ['上一个窗口', '智能预测', '最近使用', '最近使用'],
    },
    privacy: {
      title: '你的窗口，只属于你。',
      lead: '窗口画面、使用习惯和学习结果全部保存在本机，数据文件结构公开透明，可随时自查。',
      points: [
        '窗口标题、画面、文件名、应用列表绝不离开你的 Mac',
        '隐私排除名单：默认预置密码管理器与网银客户端，从不截取',
        '一键清除全部学习数据与快照缓存',
        '拒绝屏幕录制权限时，自动降级为纯标题列表，照样好用',
      ],
    },
    perf: {
      title: '快，是底线。',
      items: [
        { value: '<16 ms', label: '盲按快切延迟' },
        { value: '<50 ms', label: '按住到面板可见' },
        { value: '0%', label: '空闲时 CPU 占用' },
        { value: '<150 MB', label: '内存（含快照缓存）' },
      ],
    },
    pricing: {
      title: '免费就好用，专业版更进一步。',
      lead: '安装即享 14 天专业版全功能试用，无需绑卡。',
      free: { name: '免费版', price: '¥0', note: '永久免费', items: ['分组标签 + 中央大预览', '快切与关键词搜索', '覆盖最小化、隐藏与跨桌面窗口', '在面板内关闭、最小化、退出'] },
      pro: { name: '专业版', price: '¥68', note: '一次性买断 · 海外 $12.99', items: ['包含免费版全部功能', '第 2 顺位窗口预测', '搜索结果个性化排序', '即将推出：拼音搜索、多向布局、固定应用、多屏过滤', '1 个许可证可激活 3 台 Mac'] },
      buy: '购买专业版',
      early: '首发早鸟价 ¥48 / $9.99，支持 14 天无理由退款。',
    },
    faq: [
      { q: '为什么需要“辅助功能”权限？', a: '用于列出窗口、激活窗口，以及感知快捷键的按下与松开。macOS 上所有窗口切换器都需要这项权限。' },
      { q: '必须开启“屏幕录制”吗？', a: '不必须。它只用于显示预览；拒绝时窗签会显示纯标题列表，依然好用。' },
      { q: '会替换系统的 ⌘Tab 吗？', a: '由你决定。推荐模式保留 ⌘Tab 切换应用、用 ⌥Tab 切换窗口；全面接管模式直接用 ⌘Tab 切换窗口。' },
      { q: '支持平铺式窗口管理器吗？', a: '窗签面向 macOS 默认的堆叠式窗口，不做分屏和平铺排列。' },
      { q: '可以在 Mac App Store 下载吗？', a: '不能：窗口切换器所需的权限不被 App Store 允许。请在官网或通过 Homebrew 下载，应用会自动更新。' },
    ],
    cta: { title: '试用一天，你就回不去了。', button: '下载窗签' },
    footer: { tagline: 'Windeck 窗签 —— 你的窗口，像一副牌一样随手翻。', privacy: '隐私', contact: '联系我们' },
    demo: { search: '搜索窗口', space: '空格', hints: ['下一个', '移动', '关闭', '搜索'], minimized: '已最小化', desktop: '桌面 2', predicted: '智能预测' },
  },
} as const;

export type Content = (typeof content)['en'] | (typeof content)['zh'];
