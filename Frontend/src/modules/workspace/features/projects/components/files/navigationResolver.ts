const knownExtensions = [
  ".tsx", ".ts", ".jsx", ".js", ".mjs", ".cjs", ".json", ".css", ".scss", ".html",
  ".go", ".py", ".rs", ".java", ".md", ".yaml", ".yml", ".sql",
];

export type PathReference = {
  path: string;
  line: number | null;
  column: number | null;
};

export function pathReferenceAt(lineText: string, column: number): PathReference | null {
  const pattern = /(?:\/workspace\/|\.{1,2}[\\/])[\w.\\/-]+(?::\d+(?::\d+)?)?|[\w.\\/-]+\.(?:tsx?|jsx?|mjs|cjs|json|css|scss|html|go|py|rs|java|md|ya?ml|sql)(?::\d+(?::\d+)?)?/gi;
  const cursor = column - 1;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(lineText)) !== null) {
    if (cursor < match.index || cursor >= match.index + match[0].length) continue;
    if (/^https?:\/\//i.test(match[0]) || /^file:\/\//i.test(match[0])) continue;
    return parsePathReference(match[0]);
  }
  return null;
}

export function resolveWorkspacePathCandidates(currentPath: string, reference: PathReference): string[] {
  let pathText = reference.path.replaceAll("\\", "/");
  let segments: string[];
  if (pathText.startsWith("/workspace/")) {
    segments = [];
    pathText = pathText.slice("/workspace/".length);
  } else if (pathText.startsWith("/")) {
    return [];
  } else {
    const currentSegments = currentPath.replaceAll("\\", "/").replace(/^\/+/, "").split("/");
    if (currentSegments[0] === "workspace") currentSegments.shift();
    currentSegments.pop();
    segments = currentSegments;
  }

  for (const segment of pathText.split("/")) {
    if (!segment || segment === ".") continue;
    if (segment === "..") {
      if (segments.length === 0) return [];
      segments.pop();
    } else {
      segments.push(segment);
    }
  }
  if (segments.length === 0) return [];

  const workspacePath = `/workspace/${segments.join("/")}`;
  if (knownExtensions.some((extension) => workspacePath.toLowerCase().endsWith(extension))) {
    return [workspacePath];
  }
  return [
    workspacePath,
    ...knownExtensions.map((extension) => `${workspacePath}${extension}`),
    ...knownExtensions.map((extension) => `${workspacePath}/index${extension}`),
  ];
}

function parsePathReference(reference: string): PathReference {
  const location = /^(.*?)(?::(\d+))(?::(\d+))?$/.exec(reference);
  if (!location) return { path: reference, line: null, column: null };
  return {
    path: location[1],
    line: Number(location[2]),
    column: location[3] ? Number(location[3]) : null,
  };
}