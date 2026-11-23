<script lang="ts">
  import { navigate } from "../lib/router.svelte";
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
    </div>
  </header>

  <main class="site-main" id="main">
    <slot />
  </main>

  <footer class="site-footer">
    <div class="container">
      <p>Учебный проект команды. Проверка доступности контента.</p>
    </div>
  </footer>
</div>
