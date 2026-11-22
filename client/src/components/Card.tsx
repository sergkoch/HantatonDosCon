import type { ReactNode } from "react";

type CardProps = {
  title: string;
  children: ReactNode;
};

export function Card({ title, children }: CardProps) {
  return (
    <article className="card">
      <h3 className="card__title">{title}</h3>
      <div className="card__body">{children}</div>
    </article>
  );
}
