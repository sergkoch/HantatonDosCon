<script lang="ts">
  import type { Snippet } from "svelte";
  import { tick } from "svelte";
  import { getPath, isPlainClick, navigate } from "../lib/router.svelte";
  import { getTheme, toggleTheme } from "../lib/theme.svelte";
  import { getLocale, setLocale, t } from "../lib/i18n.svelte";
  import CookieConsent from "./CookieConsent.svelte";

  let { children }: { children: Snippet } = $props();

  const theme = $derived(getTheme());
  const path = $derived(getPath());

  const locale = $derived(getLocale());

  let langOpen = $state(false);
  let langButton: HTMLButtonElement | undefined = $state();
  let langMenu: HTMLDivElement | undefined = $state();

  const languages = [
    { value: "ru", label: "Русский", lang: "ru" },
    { value: "en", label: "English", lang: "en" },
  ] as const;

  async function toggleLangMenu() {
    langOpen = !langOpen;
    if (langOpen) {
      await tick();
      langMenu?.querySelector<HTMLButtonElement>("button")?.focus();
    }
  }

  function chooseLang(next: "ru" | "en") {
    setLocale(next);
    langOpen = false;
    langButton?.focus();
  }

  function onWindowClick(event: MouseEvent) {
    if (!langOpen) return;
    const target = event.target as HTMLElement | null;
    if (!target?.closest(".lang")) {
      langOpen = false;
    }
  }

  function onWindowKeydown(event: KeyboardEvent) {
    if (langOpen && event.key === "Escape") {
      langOpen = false;
      langButton?.focus();
    }
  }

  function onMenuKeydown(event: KeyboardEvent) {
    if (event.key !== "ArrowDown" && event.key !== "ArrowUp") {
      return;
    }

    event.preventDefault();
    const items = Array.from(
      langMenu?.querySelectorAll<HTMLButtonElement>("button") ?? [],
    );
    if (items.length === 0) return;

    const current = items.indexOf(document.activeElement as HTMLButtonElement);
    const step = event.key === "ArrowDown" ? 1 : -1;
    const next = (current + step + items.length) % items.length;
    items[next]?.focus();
  }

  $effect(() => {
    document.title = t("site.title");
    const description = document.querySelector('meta[name="description"]');
    if (description) {
      description.setAttribute("content", t("site.description"));
    }
  });

  const tactilePaths = [
    "0,12 9,12 9,26 3,26 3,34 0,34",
    "0,58 8,58 8,68 0,68",
    "100,15 91,15 91,29 97,29 97,37 100,37",
    "100,62 92,62 92,72 100,72",
  ];
</script>

<svelte:window onclick={onWindowClick} onkeydown={onWindowKeydown} />

<div class="site">
  <a class="skip-link" href="#main">{t("layout.skip")}</a>

  <header class="site-header">
    <div class="container site-header__inner">
      <a
        class="brand"
        href="/"
        onclick={(event) => {
          if (!isPlainClick(event)) return;
          event.preventDefault();
          navigate("/");
        }}
      >
        <span class="brand__mark" aria-hidden="true">
          <img src="/logo.png" alt="" />
        </span>
        {t("site.title")}
      </a>

      <div class="site-header__nav">
        <nav class="site-nav" aria-label={t("layout.navLabel")}>
          <a
            href="/"
            aria-current={path === "/" ? "page" : undefined}
            onclick={(event) => {
              if (!isPlainClick(event)) return;
              event.preventDefault();
              navigate("/");
            }}
          >
            {t("layout.navCheck")}
          </a>
          <a
            href="/history"
            aria-current={path === "/history" ? "page" : undefined}
            onclick={(event) => {
              if (!isPlainClick(event)) return;
              event.preventDefault();
              navigate("/history");
            }}
          >
            {t("layout.navHistory")}
          </a>
          <a href="#how">{t("layout.navHow")}</a>
        </nav>

        <div class="lang">
          <button
            bind:this={langButton}
            class="lang-button"
            type="button"
            aria-haspopup="menu"
            aria-expanded={langOpen}
            aria-controls="lang-menu"
            onclick={toggleLangMenu}
          >
            {t("layout.langButton")}
            <svg
              class={"lang-button__chevron" + (langOpen ? " is-open" : "")}
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              aria-hidden="true"
            >
              <path
                d="m6 9 6 6 6-6"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>

          {#if langOpen}
            <div
              bind:this={langMenu}
              class="lang-menu"
              id="lang-menu"
              role="menu"
              tabindex="-1"
              aria-label={t("layout.langMenu")}
              onkeydown={onMenuKeydown}
            >
              {#each languages as item (item.value)}
                <button
                  class="lang-menu__item"
                  type="button"
                  role="menuitemradio"
                  aria-checked={locale === item.value}
                  onclick={() => chooseLang(item.value)}
                >
                  <span lang={item.lang}>{item.label}</span>
                  {#if locale === item.value}
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      aria-hidden="true"
                    >
                      <path
                        d="m5 13 4 4L19 7"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
        </div>

        <button
          class="theme-toggle"
          type="button"
          onclick={(event) => toggleTheme(event)}
          aria-label={theme === "dark"
            ? t("layout.themeToLight")
            : t("layout.themeToDark")}
          title={theme === "dark"
            ? t("layout.themeLightTitle")
            : t("layout.themeDarkTitle")}
        >
          {#if theme === "dark"}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle
                cx="12"
                cy="12"
                r="4"
                stroke="currentColor"
                stroke-width="1.8"
              />
              <path
                d="M12 2v2M12 20v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M2 12h2M20 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
              />
            </svg>
          {:else}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path
                d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linejoin="round"
              />
            </svg>
          {/if}
        </button>
      </div>
    </div>
  </header>

  <main class="site-main" id="main" tabindex="-1">
    {@render children()}
  </main>

  <footer class="site-footer">
    <div class="container">
      <p class="site-footer__contacts">
        Контакты команды:
        <a href="mailto:Gayka.pon@mail.ru">Gayka.pon@mail.ru</a>
        <span aria-hidden="true">·</span>
        <a href="tel:+79825457103">+7 982 545-71-03</a>
      </p>
    </div>
  </footer>

  <CookieConsent />

  <svg
    class="tactile-lines"
    viewBox="0 0 100 100"
    preserveAspectRatio="none"
    aria-hidden="true"
  >
    {#each tactilePaths as points (points)}
      <polyline class="tl-base" {points} />
      <polyline class="tl-mid" {points} />
      <polyline class="tl-core" {points} />
      <polyline class="tl-dark" {points} />
    {/each}
  </svg>

  <div class="tactile-dots" aria-hidden="true">
    <span class="tactile-corner" style="left:9%; top:12%;"></span>
    <span class="tactile-corner" style="left:9%; top:26%;"></span>
    <span class="tactile-corner" style="left:3%; top:26%;"></span>
    <span class="tactile-corner" style="left:3%; top:34%;"></span>
    <span class="tactile-corner" style="left:8%; top:58%;"></span>
    <span class="tactile-corner" style="left:8%; top:68%;"></span>
    <span class="tactile-corner" style="left:91%; top:15%;"></span>
    <span class="tactile-corner" style="left:91%; top:29%;"></span>
    <span class="tactile-corner" style="left:97%; top:29%;"></span>
    <span class="tactile-corner" style="left:97%; top:37%;"></span>
    <span class="tactile-corner" style="left:92%; top:62%;"></span>
    <span class="tactile-corner" style="left:92%; top:72%;"></span>
  </div>
</div>
