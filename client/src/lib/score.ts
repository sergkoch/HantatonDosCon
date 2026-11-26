import type { FindingKind } from "../types";
import { type Locale, translate } from "./messages";

export function scoreLabel(score: number | null, locale: Locale = "ru"): string {
  if (score === null) {
    return translate(locale, "score.inProgress");
  }

  if (score >= 80) {
    return translate(locale, "score.good");
  }

  if (score >= 50) {
    return translate(locale, "score.needsWork");
  }

  return translate(locale, "score.critical");
}

const kindKeys = {
  image: "kind.image",
  pdf: "kind.pdf",
  link: "kind.link",
  page: "kind.page",
  form: "kind.form",
  table: "kind.table",
} as const satisfies Record<FindingKind, string>;

export function kindLabel(kind: FindingKind, locale: Locale = "ru"): string {
  return translate(locale, kindKeys[kind]);
}
