import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="container">
      <section className="section">
        <h1 className="section__title">Страница не найдена</h1>
        <p>Проверьте адрес или вернитесь на главную.</p>
        <Link className="button button--primary" to="/">
          На главную
        </Link>
      </section>
    </div>
  );
}
