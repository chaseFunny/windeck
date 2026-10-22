// The site's words, in English and Simplified Chinese: one page, two
// languages, the same layout. Facts only: every number here is measured
// or set, never a goal.

export type Lang = 'en' | 'zh';

export const SITE_URL = 'https://windeck.islumi.com';
export const REPO_URL = 'https://github.com/chaseFunny/windeck';
// The latest release's disk image, under a name that does not change.
export const DOWNLOAD_URL = `${REPO_URL}/releases/latest/download/Windeck-arm64.dmg`;
export const RELEASES_URL = `${REPO_URL}/releases`;

export const content = {
  en: {
    nav: { features: 'Features', smart: 'Smart', privacy: 'Privacy', pricing: 'Pricing', faq: 'FAQ', download: 'Download', other: '中文', otherHref: '/zh' },
    hero: {
      eyebrow: 'A window switcher for macOS',
      title: 'Switch windows like browser tabs.',
      lead: 'Hold ⌥ and press Tab. Every window shows as a readable tab, grouped by app, with a large preview to be sure. Let go, and you are there.',
      download: 'Download for macOS',
      how: 'See how it works',
      note: 'Free · macOS 13 or later · Apple silicon',
    },
    stepsTitle: 'Three moves, no more.',
    steps: [
      { keys: ['⌥', '⇥'], title: 'Tap to go back', body: 'A quick press and release returns to the previous window. Nothing flashes on screen.' },
      { keys: ['⌥', '⇥', '⇥'], title: 'Hold to look', body: 'Keep holding and the tabs appear. Tab moves along your recent windows, arrows along the list.' },
      { keys: ['⌥', '␣'], title: 'Type to find', body: 'Press Space and let go of ⌥. Type a few letters of any title; Return switches.' },
    ],
    featuresTitle: 'Find it fast. Be sure it is the one.',
    featuresLead: 'Tabs do the finding, the preview does the confirming.',
    features: [
      { title: 'Full titles, grouped by app', body: 'A vertical tab bar keeps every title readable; apps with several windows fold into a group.' },
      { title: 'A large, sharp preview', body: 'The highlighted window shows big, so two similar editor windows never get mixed up.' },
      { title: 'Every window, wherever it is', body: 'Minimized, hidden, on another desktop or display: all listed, each with a badge saying where.' },
      { title: 'Places that stay put', body: 'Apps and windows keep their place in the list; only the highlight moves.' },
      { title: 'Act without leaving', body: 'W closes, M minimizes, Q quits the app, while the panel stays. Rather not? Turn them off in Settings → Shortcuts.' },
      { title: 'Light and native', body: 'One native app of a few megabytes, driven by events rather than polling, that keeps itself up to date.' },
    ],
    smart: {
      title: 'Smart where it helps. Never in the way.',
      lead: 'Windeck learns which window you go to next, on your Mac only, and never touches what your fingers rely on.',
      points: [
        { title: 'The first Tab is always the previous window', body: 'No prediction ever takes the first place. A blind tap goes back, every time.' },
        { title: 'The second place is predicted', body: 'The window you usually go to next waits one Tab further, marked with a small star.' },
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
      list: ['1Password', 'Keychain Access', 'Passwords', 'Bitwarden'],
      listTitle: 'Privacy list',
      never: 'Never captured',
    },
    facts: {
      title: 'Small, quiet, local.',
      items: [
        { value: '4.8 MB', label: 'to download' },
        { value: '0', label: 'window titles sent anywhere' },
        { value: '13+', label: 'macOS version' },
        { value: '2', label: 'languages: English, 简体中文' },
      ],
    },
    pricing: {
      title: 'Free to switch. Pro to go further.',
      lead: 'Every install starts with a 14-day Pro trial. No card needed.',
      free: { name: 'Free', price: '$0', note: 'forever', items: ['Grouped tabs and large previews', 'Quick switch and keyword search', 'Minimized, hidden and other-desktop windows', 'Close, minimize and quit from the panel'] },
      pro: { name: 'Pro', price: '$12.99', note: 'once · ¥68 in China', items: ['Everything in Free', 'Predicted second window', 'Search ranked by your habits', 'One license, three Macs'] },
      buy: 'Buying opens soon',
      early: 'Pro can be bought soon; until then, every install has 14 days of Pro.',
    },
    faqTitle: 'Questions',
    faq: [
      { q: 'Why does Windeck need Accessibility?', a: 'To list your windows, bring one forward, and hear the shortcut being held and released. macOS asks for it for every window switcher.' },
      { q: 'Is Screen Recording required?', a: 'No. It is only used to show previews. Without it, Windeck shows titles alone, which still works well.' },
      { q: 'Does it replace ⌘Tab?', a: 'Your choice. The recommended mode keeps ⌘Tab for apps and uses ⌥Tab for windows; the takeover mode uses ⌘Tab for windows.' },
      { q: 'Can it show just icons and titles?', a: 'Yes. Choose the List layout in Settings → General, or from the menu bar: one column of app icons and window titles, the most recent first.' },
      { q: 'How does Windeck update?', a: 'It checks GitHub Releases once a day, shows what is new and installs the update when you agree, or on its own if you choose so. Updates are signed, and Windeck installs only those signed with its key. Check any time from the menu bar.' },
      { q: 'What about tiling window managers?', a: 'Windeck is for the usual stacked windows of macOS. It neither tiles nor arranges windows.' },
      { q: 'macOS says it cannot check Windeck. What now?', a: 'This first version is not notarized by Apple yet. Open System Settings → Privacy & Security and click Open Anyway, once.' },
      { q: 'Is it on the Mac App Store?', a: 'No: the permissions a window switcher needs are not allowed there. Download it from this page or from GitHub Releases.' },
    ],
    cta: { title: 'Try it for a day. Keep it for years.', lead: 'Free to download, with every Pro feature for 14 days.', button: 'Download Windeck' },
    footer: { tagline: 'Windeck · 窗签 — your windows, at hand like a deck of cards.', privacy: 'Privacy', contact: 'Contact', made: 'Made by' },
    demo: { search: 'Search windows', space: 'Space', hints: ['Next', 'Close', 'Minimize', 'Search'], minimized: 'Minimized', desktop: 'Desktop 2', predicted: 'Predicted' },
  },
  zh: {
    nav: { features: '功能', smart: '智能', privacy: '隐私', pricing: '价格', faq: '常见问题', download: '下载', other: 'English', otherHref: '/' },
    hero: {
      eyebrow: 'macOS 窗口切换器',
      title: '像切浏览器标签页一样切换窗口',
      lead: '按住 ⌥ 再按 Tab，所有窗口按应用分组、以完整标题的标签呈现，中央大预览帮你确认“就是它”。松开即到达。',
      download: '下载 macOS 版',
      how: '看看怎么用',
      note: '免费 · macOS 13 及以上 · Apple 芯片',
    },
    stepsTitle: '三个动作，就够了。',
    steps: [
      { keys: ['⌥', '⇥'], title: '轻按，回到上一个窗口', body: '快速按下并松开，直接回到刚才的窗口，屏幕上什么都不会闪。' },
      { keys: ['⌥', '⇥', '⇥'], title: '按住，看清所有窗口', body: '继续按住，标签栏展开。Tab 按最近使用顺序跳转，方向键按列表上下移动。' },
      { keys: ['⌥', '␣'], title: '打字，找到它', body: '按空格后即可松开 ⌥，输入标题里的几个字，回车即切换。' },
    ],
    featuresTitle: '找得快，也认得准。',
    featuresLead: '标签负责找，预览负责确认。',
    features: [
      { title: '完整标题，按应用分组', body: '左侧纵向标签栏让每个标题完整可读，多窗口的应用折叠为分组。' },
      { title: '高清大预览', body: '选中的窗口在中央大图显示，两个相似的编辑器窗口再也不会认错。' },
      { title: '不管窗口在哪', body: '最小化、已隐藏、在其他桌面或外接显示器的窗口全部列出，并用徽标标明位置。' },
      { title: '位置稳定不乱跳', body: '应用与窗口在列表中的位置固定，只有高亮在移动。' },
      { title: '不离开面板就能处理', body: 'W 关闭、M 最小化、Q 退出应用，面板保持不动；不需要的话，可在 设置 → 快捷键 里关掉。' },
      { title: '轻巧、原生', body: '只有几 MB 的原生应用，靠系统事件驱动，不轮询，还会自动更新。' },
    ],
    smart: {
      title: '智能只在需要时出现，从不碍事。',
      lead: '窗签在本机学习你接下来常去的窗口，但绝不改动你的肌肉记忆。',
      points: [
        { title: '第 1 顺位永远是上一个窗口', body: '任何预测都不会占用第 1 顺位，盲按一下 Tab 永远回到刚才的窗口。' },
        { title: '第 2 顺位智能预测', body: '你最可能要去的窗口就在下一个 Tab，并带有小星标提示。' },
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
      list: ['1Password', '钥匙串访问', '密码', 'Bitwarden'],
      listTitle: '隐私排除名单',
      never: '从不截取',
    },
    facts: {
      title: '小巧、安静、只在本机。',
      items: [
        { value: '4.8 MB', label: '安装包大小' },
        { value: '0', label: '条窗口标题被上传' },
        { value: '13+', label: '支持的 macOS 版本' },
        { value: '2', label: '种语言：English、简体中文' },
      ],
    },
    pricing: {
      title: '免费就好用，专业版更进一步。',
      lead: '安装即享 14 天专业版全功能试用，无需绑卡。',
      free: { name: '免费版', price: '¥0', note: '永久免费', items: ['分组标签 + 中央大预览', '快切与关键词搜索', '覆盖最小化、隐藏与跨桌面窗口', '在面板内关闭、最小化、退出'] },
      pro: { name: '专业版', price: '¥68', note: '一次性买断 · 海外 $12.99', items: ['包含免费版全部功能', '第 2 顺位窗口预测', '搜索结果个性化排序', '1 个许可证可激活 3 台 Mac'] },
      buy: '即将开放购买',
      early: '专业版即将开放购买；在此之前，每次安装都可免费使用 14 天专业版。',
    },
    faqTitle: '常见问题',
    faq: [
      { q: '为什么需要“辅助功能”权限？', a: '用于列出窗口、激活窗口，以及感知快捷键的按下与松开。macOS 上所有窗口切换器都需要这项权限。' },
      { q: '必须开启“屏幕录制”吗？', a: '不必须。它只用于显示预览；拒绝时窗签会显示纯标题列表，依然好用。' },
      { q: '会替换系统的 ⌘Tab 吗？', a: '由你决定。推荐模式保留 ⌘Tab 切换应用、用 ⌥Tab 切换窗口；全面接管模式直接用 ⌘Tab 切换窗口。' },
      { q: '可以只显示图标和标题吗？', a: '可以。在 设置 → 通用 或菜单栏里选择“列表”布局：一列应用图标加窗口标题，最近用过的排在前面。' },
      { q: '窗签怎么更新？', a: '每天检查一次 GitHub Releases，有新版本时展示更新内容，你同意后安装；也可以设置为自动安装。更新包带签名，窗签只安装用自己密钥签名的版本。也可以随时在菜单栏点“检查更新”。' },
      { q: '支持平铺式窗口管理器吗？', a: '窗签面向 macOS 默认的堆叠式窗口，不做分屏和平铺排列。' },
      { q: '打开时提示“无法验证开发者”怎么办？', a: '首个版本还没有经过 Apple 公证。打开 系统设置 → 隐私与安全性，点“仍要打开”即可，只需一次。' },
      { q: '可以在 Mac App Store 下载吗？', a: '不能：窗口切换器所需的权限不被 App Store 允许。请在官网或 GitHub Releases 下载。' },
    ],
    cta: { title: '试用一天，你就回不去了。', lead: '免费下载，前 14 天专业版功能全部可用。', button: '下载窗签' },
    footer: { tagline: 'Windeck 窗签 —— 你的窗口，像一副牌一样随手翻。', privacy: '隐私', contact: '联系我们', made: '出品' },
    demo: { search: '搜索窗口', space: '空格', hints: ['下一个', '关闭', '最小化', '搜索'], minimized: '已最小化', desktop: '桌面 2', predicted: '智能预测' },
  },
} as const;

export type Content = (typeof content)['en'] | (typeof content)['zh'];
