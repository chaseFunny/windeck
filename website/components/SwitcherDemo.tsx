import { For, Show, createSignal, onCleanup, onMount } from 'solid-js';
import type { Content } from '../lib/content';

// A live model of the switcher: the grouped tabs at the left, the large
// preview in the middle, the highlight walking the cycle as Tab would.
// The apps are made up, drawn in CSS.

type Kind = 'editor' | 'terminal' | 'browser' | 'doc' | 'design';

interface Win {
  id: number;
  title: string;
  kind: Kind;
  badge?: 'minimized' | 'desktop';
}

interface Group {
  app: string;
  color: string;
  letter: string;
  windows: Win[];
}

const groups: Group[] = [
  {
    app: 'Code Editor',
    color: 'linear-gradient(160deg,#4aa3ff,#1e6fe8)',
    letter: '{ }',
    windows: [
      { id: 1, title: 'server.ts — api', kind: 'editor' },
      { id: 2, title: 'App.tsx — web', kind: 'editor' },
    ],
  },
  { app: 'Terminal', color: 'linear-gradient(160deg,#4a4a52,#1c1c20)', letter: '>_', windows: [{ id: 3, title: 'zsh — ~/projects/api', kind: 'terminal' }] },
  {
    app: 'Browser',
    color: 'linear-gradient(160deg,#5ad0ff,#2a7bff)',
    letter: '◎',
    windows: [
      { id: 4, title: 'localhost:5173 — Dashboard', kind: 'browser' },
      { id: 5, title: 'net/http — Go Packages', kind: 'browser' },
    ],
  },
  { app: 'Notes', color: 'linear-gradient(160deg,#ffd76a,#f5a524)', letter: '✎', windows: [{ id: 6, title: 'Q3 Roadmap', kind: 'doc', badge: 'minimized' }] },
  { app: 'Design', color: 'linear-gradient(160deg,#c084fc,#7c3aed)', letter: '◆', windows: [{ id: 7, title: 'Landing page — v4', kind: 'design', badge: 'desktop' }] },
];

// Tab's order: the previous window, the predicted one, then recency.
const cycle = [3, 4, 2, 5, 6, 7];
const predicted = 4;

const all = groups.flatMap((g) => g.windows.map((w) => ({ ...w, group: g })));
const byId = (id: number) => all.find((w) => w.id === id)!;

function AppIcon(props: { group: Group; size?: number }) {
  const s = () => props.size ?? 22;
  return (
    <span class="demo-icon" style={{ width: `${s()}px`, height: `${s()}px`, background: props.group.color, 'font-size': `${s() * 0.42}px` }}>
      {props.group.letter}
    </span>
  );
}

function Preview(props: { kind: Kind }) {
  return (
    <div class={`fake fake--${props.kind}`}>
      <div class="fake__bar">
        <i />
        <i />
        <i />
        <Show when={props.kind === 'browser'}>
          <span class="fake__url" />
        </Show>
      </div>
      <div class="fake__body">
        <Show when={props.kind === 'editor'}>
          <div class="fake__side" />
          <div class="fake__code">
            <For each={[62, 40, 75, 55, 30, 68, 48, 80, 36, 58, 44]}>
              {(w, i) => <span style={{ width: `${w}%`, 'margin-left': `${(i() % 3) * 6}%` }} />}
            </For>
          </div>
        </Show>
        <Show when={props.kind === 'terminal'}>
          <div class="fake__term">
            <For each={['$ bun run test', '✓ 128 passed', '$ go build ./...', '$ ▍']}>{(l) => <p>{l}</p>}</For>
          </div>
        </Show>
        <Show when={props.kind === 'browser'}>
          <div class="fake__page">
            <div class="fake__hero" />
            <div class="fake__cards">
              <i />
              <i />
              <i />
            </div>
          </div>
        </Show>
        <Show when={props.kind === 'doc'}>
          <div class="fake__doc">
            <b />
            <For each={[90, 84, 70, 88, 60]}>{(w) => <span style={{ width: `${w}%` }} />}</For>
          </div>
        </Show>
        <Show when={props.kind === 'design'}>
          <div class="fake__canvas">
            <i />
            <i />
          </div>
        </Show>
      </div>
    </div>
  );
}

export default function SwitcherDemo(props: { t: Content }) {
  const [sel, setSel] = createSignal(cycle[0]);
  const [press, setPress] = createSignal(false);
  let paused = false;
  let step = 0;
  let timer: ReturnType<typeof setInterval> | undefined;
  onMount(() => {
    timer = setInterval(() => {
      if (paused) return;
      step = (step + 1) % cycle.length;
      setSel(cycle[step]);
      setPress(true);
      setTimeout(() => setPress(false), 220);
    }, 1700);
  });
  onCleanup(() => clearInterval(timer));
  const current = () => byId(sel());
  const ordinal = (id: number) => {
    const i = cycle.indexOf(id);
    return i >= 0 && i < 3 ? i + 1 : 0;
  };
  const t = () => props.t.demo;

  return (
    <div class="demo" onMouseEnter={() => (paused = true)} onMouseLeave={() => (paused = false)}>
      <div class="demo__panel" role="img" aria-label="Windeck switcher">
        <aside class="demo__side">
          <div class="demo__search">
            <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="7" cy="7" r="4.5" />
              <path d="M10.4 10.4 14 14" />
            </svg>
            <span>{t().search}</span>
            <kbd>{t().space}</kbd>
          </div>
          <div class="demo__list">
            <For each={groups}>
              {(g) => (
                <>
                  <Show when={g.windows.length > 1}>
                    <div class="demo__group">
                      <span class="demo__chev">⌄</span>
                      <AppIcon group={g} size={16} />
                      <span>{g.app}</span>
                      <em>{g.windows.length}</em>
                    </div>
                  </Show>
                  <For each={g.windows}>
                    {(w) => (
                      <div
                        class="demo__row"
                        classList={{ 'is-sel': sel() === w.id, 'is-child': g.windows.length > 1 }}
                        onMouseEnter={() => setSel(w.id)}
                      >
                        {g.windows.length > 1 ? <span class="demo__guide" /> : <AppIcon group={g} />}
                        <span class="demo__title">
                          {w.title}
                          <Show when={g.windows.length === 1}>
                            <small>{g.app}</small>
                          </Show>
                        </span>
                        <Show when={w.badge}>
                          <span class="demo__badge">{w.badge === 'minimized' ? t().minimized : t().desktop}</span>
                        </Show>
                        <Show when={w.id === predicted}>
                          <span class="demo__star" title={t().predicted}>★</span>
                        </Show>
                        <Show when={ordinal(w.id)}>
                          <span class="demo__ord">{ordinal(w.id)}</span>
                        </Show>
                      </div>
                    )}
                  </For>
                </>
              )}
            </For>
          </div>
          <div class="demo__hints">
            <span><kbd>⇥</kbd>{t().hints[0]}</span>
            <span><kbd>↑↓</kbd>{t().hints[1]}</span>
            <span><kbd>W</kbd>{t().hints[2]}</span>
            <span><kbd>␣</kbd>{t().hints[3]}</span>
          </div>
        </aside>
        <section class="demo__stage">
          <header class="demo__head">
            <AppIcon group={current().group} size={30} />
            <div>
              <strong>{current().title}</strong>
              <small>{current().group.app}</small>
            </div>
          </header>
          <div class="demo__preview">
            <Preview kind={current().kind} />
          </div>
        </section>
      </div>
      <div class="demo__keys" aria-hidden="true">
        <kbd class="key is-held">⌥ option</kbd>
        <span>+</span>
        <kbd class="key" classList={{ 'is-down': press() }}>⇥ tab</kbd>
      </div>
    </div>
  );
}
