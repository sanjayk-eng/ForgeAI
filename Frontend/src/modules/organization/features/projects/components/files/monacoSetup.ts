import { loader } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import cssWorker from "monaco-editor/language/css/css.worker?worker";
import editorWorker from "monaco-editor/editor/editor.worker?worker";
import htmlWorker from "monaco-editor/language/html/html.worker?worker";
import jsonWorker from "monaco-editor/language/json/json.worker?worker";
import typescriptWorker from "monaco-editor/language/typescript/ts.worker?worker";

monaco.editor.defineTheme("forge-dark", {
  base: "vs-dark",
  inherit: true,
  rules: [
    { token: "comment", foreground: "899AB3", fontStyle: "italic" },
    { token: "keyword", foreground: "68A8FF" },
    { token: "string", foreground: "9BD88F" },
    { token: "number", foreground: "E8AD69" },
    { token: "type", foreground: "79D2C5" },
  ],
  colors: {
    "editor.background": "#0C111A",
    "editor.foreground": "#EAF0FA",
    "editorLineNumber.foreground": "#596D88",
    "editorLineNumber.activeForeground": "#68A8FF",
    "editorCursor.foreground": "#68A8FF",
    "editor.selectionBackground": "#418BE855",
    "editor.lineHighlightBackground": "#FFFFFF08",
    "editorIndentGuide.background1": "#B8CCEB20",
    "editorGutter.background": "#0C111A",
    "editorWidget.background": "#151D2A",
    "editorWidget.border": "#B8CCEB26",
    focusBorder: "#68A8FF",
  },
});

monaco.editor.defineTheme("forge-light", {
  base: "vs",
  inherit: true,
  rules: [
    { token: "comment", foreground: "596D88", fontStyle: "italic" },
    { token: "keyword", foreground: "1D4ED8" },
    { token: "string", foreground: "18794E" },
    { token: "number", foreground: "A85F1B" },
    { token: "type", foreground: "6D45B5" },
  ],
  colors: {
    "editor.background": "#EEF3FA",
    "editor.foreground": "#172234",
    "editorLineNumber.foreground": "#71829A",
    "editorLineNumber.activeForeground": "#2563EB",
    "editorCursor.foreground": "#2563EB",
    "editor.selectionBackground": "#2563EB33",
    "editor.lineHighlightBackground": "#17223408",
    "editorIndentGuide.background1": "#1722341A",
    "editorGutter.background": "#EEF3FA",
    "editorWidget.background": "#FFFFFF",
    "editorWidget.border": "#17223429",
    focusBorder: "#2563EB",
  },
});

globalThis.MonacoEnvironment = {
  getWorker(_workerId, label) {
    if (label === "json") return new jsonWorker();
    if (["css", "scss", "less"].includes(label)) return new cssWorker();
    if (["html", "handlebars", "razor"].includes(label)) return new htmlWorker();
    if (["typescript", "javascript"].includes(label)) return new typescriptWorker();
    return new editorWorker();
  },
};

loader.config({ monaco });