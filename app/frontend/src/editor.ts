import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags } from "@lezer/highlight";

/**
 * 编辑器配色与首页演示台用同一套 token 颜色：关键字深青、字符串赭黄、
 * 数字与主色同族、注释浅灰。默认高亮是蓝紫红，放在这套克制的界面里太吵。
 */
const goHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: "#0a3a46" },
  { tag: tags.typeName, color: "#0a3a46" },
  { tag: tags.bool, color: "#0a3a46" },
  { tag: tags.null, color: "#0a3a46" },
  { tag: tags.function(tags.variableName), color: "#11181b" },
  { tag: tags.definition(tags.variableName), color: "#11181b" },
  { tag: tags.string, color: "#7a4f08" },
  { tag: tags.number, color: "#0f5666" },
  { tag: tags.comment, color: "#9aa3a5" },
  { tag: tags.operator, color: "#6d7779" },
  { tag: tags.punctuation, color: "#6d7779" },
  { tag: tags.bracket, color: "#6d7779" },
]);

export const editorTheme = [syntaxHighlighting(goHighlight)];
