import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags } from "@lezer/highlight";

/**
 * 编辑器配色与首页演示台共用一套 token 色值，都跑在墨色表面上：
 * 关键字亮青、字符串赭黄、数字薄荷绿、注释灰。默认高亮是蓝紫红，太吵。
 */
const goHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: "#5fc9db" },
  { tag: tags.typeName, color: "#5fc9db" },
  { tag: tags.bool, color: "#5fc9db" },
  { tag: tags.null, color: "#5fc9db" },
  { tag: tags.function(tags.variableName), color: "#c6d5d9" },
  { tag: tags.definition(tags.variableName), color: "#dfe8ea" },
  { tag: tags.string, color: "#e0b268" },
  { tag: tags.number, color: "#7ed0bf" },
  { tag: tags.comment, color: "#6b7d83" },
  { tag: tags.operator, color: "#9aabb0" },
  { tag: tags.punctuation, color: "#9aabb0" },
  { tag: tags.bracket, color: "#9aabb0" },
]);

export const editorTheme = [syntaxHighlighting(goHighlight)];
