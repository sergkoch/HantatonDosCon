import type { FindingKind } from "../types";

export function scoreLabel(score: number | null): string {
  if (score === null) {
    return "Проверка ещё идёт";
  }

  if (score >= 80) {
    return "Хорошо";
  }

  if (score >= 50) {
    return "Нужно исправить";
  }

  return "Критично";
}

const kindLabels: Record<FindingKind, string> = {
  image: "Изображение",
  pdf: "PDF",
  link: "Ссылка",
};

export function kindLabel(kind: FindingKind): string {
  return kindLabels[kind];
}
