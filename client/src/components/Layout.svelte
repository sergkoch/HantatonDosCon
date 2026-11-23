<script lang="ts">
  import type { Snippet } from "svelte";
  import { navigate } from "../lib/router.svelte";
  import { getTheme, toggleTheme } from "../lib/theme.svelte";

  let { children }: { children: Snippet } = $props();

  const theme = $derived(getTheme());
</script>

<div class="site">
  <a class="skip-link" href="#main">К содержимому</a>

  <header class="site-header">
    <div class="container site-header__inner">
      <a
        class="brand"
        href="/"
        onclick={(event) => {
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
            onclick={(event) => {
              event.preventDefault();
              navigate("/");
            }}
          >
            Проверить
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

  <main class="site-main" id="main">
    {@render children()}
  </main>

  <footer class="site-footer">
    <div class="container">
      <p>Учебный проект команды. Проверка доступности контента.</p>
    </div>
  </footer>
</div>
