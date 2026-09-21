import {
  Archive,
  Box,
  Code2,
  Database,
  Folder,
  GitBranch,
  Globe,
  Layers3,
  LayoutDashboard,
  Rocket,
  Settings2,
  SquareTerminal,
  Users,
  type LucideIcon,
} from "lucide-react";

export type WorkspaceTab = "preview" | "code" | "components" | "console";
export type DeviceMode = "desktop" | "tablet" | "mobile";
export type MessageRole = "user" | "agent";
export type ChatMessage = { role: MessageRole; text: string };
export type WorkspaceFile = { name: string; language: string; code: string };
export type NavItem = { label: string; icon: LucideIcon };

export const workspaceSections: { title: string; items: NavItem[] }[] = [
  {
    title: "Workspace",
    items: [
      { label: "Overview", icon: LayoutDashboard },
      { label: "Projects", icon: Box },
      { label: "Templates", icon: Layers3 },
      { label: "Assets", icon: Archive },
      { label: "Team", icon: Users },
    ],
  },
  {
    title: "Development",
    items: [
      { label: "Files", icon: Folder },
      { label: "Database", icon: Database },
      { label: "API", icon: Code2 },
      { label: "Environment", icon: Settings2 },
      { label: "GitHub", icon: GitBranch },
    ],
  },
  {
    title: "Deployment",
    items: [
      { label: "Deployments", icon: Rocket },
      { label: "Domains", icon: Globe },
      { label: "Logs", icon: SquareTerminal },
    ],
  },
];

export const projectFiles = [
  "src/App.tsx",
  "src/pages/Home.tsx",
  "src/components/Navbar.tsx",
  "package.json",
  "README.md",
];

export const homeCode = `export default function Home() {
  return (
    <main className="min-h-screen">
      <Navbar />
      <Hero />
      <FeatureGrid />
    </main>
  );
}`;

export const iconButtonClass =
  "grid size-8 place-items-center rounded-md border border-transparent text-forge-muted transition hover:border-white/10 hover:bg-white/[0.05] hover:text-[#f1f3f5]";
export const smallButtonClass =
  "inline-flex items-center gap-2 rounded-md border border-white/10 px-3 py-2 text-xs font-semibold text-forge-soft transition hover:border-white/20 hover:bg-white/[0.05] hover:text-[#f1f3f5]";
