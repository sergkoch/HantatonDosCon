import { Link, Outlet } from "react-router-dom";

export function Layout() {
  return (
    <div className="site">
      <a className="skip-link" href="#main">
        К содержимому
      </a>

      <header className="site-header">
        <div className="container site-header__inner">
          <Link className="brand" to="/">
            Доступный контент
          </Link>

          <nav className="site-nav" aria-label="Основная навигация">
            <Link to="/">Проверить</Link>
            <Link to="/#how">Как это работает</Link>
          </nav>
        </div>
      </header>

      <main className="site-main" id="main">
        <Outlet />
      </main>

      <footer className="site-footer">
        <div className="container">
          <p>Учебный проект команды. Проверка доступности контента.</p>
        </div>
      </footer>
    </div>
  );
}
