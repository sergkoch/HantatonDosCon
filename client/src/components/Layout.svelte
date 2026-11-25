<script lang="ts">
  import type { Snippet } from "svelte";
  import { getPath, isPlainClick, navigate } from "../lib/router.svelte";
  import { getTheme, toggleTheme } from "../lib/theme.svelte";

  let { children }: { children: Snippet } = $props();

  const theme = $derived(getTheme());
  const path = $derived(getPath());

  const tactilePaths = [
    "0,12 9,12 9,26 3,26 3,34 0,34",
    "0,58 8,58 8,68 0,68",
    "100,15 91,15 91,29 97,29 97,37 100,37",
    "100,62 92,62 92,72 100,72",
  ];
</script>

<div class="site">
  <a class="skip-link" href="#main">К содержимому</a>

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
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path
              d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <circle
              cx="12"
              cy="12"
              r="3"
              stroke="currentColor"
              stroke-width="1.8"
            />
          </svg>
        </span>
        Доступный контент
      </a>

      <div class="site-header__nav">
        <nav class="site-nav" aria-label="Основная навигация">
          <a
            href="/"
            aria-current={path === "/" ? "page" : undefined}
            onclick={(event) => {
              if (!isPlainClick(event)) return;
              event.preventDefault();
              navigate("/");
            }}
          >
            Проверить
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
            История
          </a>
          <a href="#how">Как это работает</a>
        </nav>

        <button
          class="theme-toggle"
          type="button"
          onclick={(event) => toggleTheme(event)}
          aria-label={theme === "dark"
            ? "Включить светлую тему"
            : "Включить тёмную тему"}
          title={theme === "dark" ? "Светлая тема" : "Тёмная тема"}
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
      <p>Учебный проект команды. Проверка доступности контента.</p>
    </div>
  </footer>

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
