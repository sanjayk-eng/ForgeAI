import { useState, type CSSProperties } from "react";
import { useQuery } from "@tanstack/react-query";
import { getCurrentUser } from "../auth/api";
import { useAuth } from "../auth/useAuth";
import {
  AssistantPanel,
  CodePanel,
  ComponentsPanel,
  ConsolePanel,
  PreviewPanel,
} from "./WorkspacePanels";
import {
  MobileNav,
  TopBar,
  WorkspaceHeader,
  WorkspaceSidebar,
  WorkspaceTabs,
} from "./WorkspaceChrome";
import type { ChatMessage, DeviceMode, WorkspaceTab } from "./workspaceData";

export function WorkspacePage() {
  const { user, tokens, signOut } = useAuth();
  const { data: profile } = useQuery({
    queryKey: ["auth", "me", tokens?.access_token],
    queryFn: () => getCurrentUser(tokens!.access_token),
    enabled: Boolean(tokens?.access_token),
  });
  const currentUser = profile ?? user;
  const name = currentUser?.name?.trim() || "Workspace user";
  const email = currentUser?.email?.trim() || "No email available";
  const initials = name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [activeNav, setActiveNav] = useState("Overview");
  const [activeTab, setActiveTab] = useState<WorkspaceTab>("preview");
  const [device, setDevice] = useState<DeviceMode>("desktop");
  const [assistantOpen, setAssistantOpen] = useState(true);
  const [prompt, setPrompt] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const workspaceColumns = `${sidebarOpen ? "224px" : "0px"} minmax(0, 1fr) ${assistantOpen ? "330px" : "0px"}`;

  function sendPrompt() {
    const value = prompt.trim();
    if (!value) return;
    setMessages((current) => [
      ...current,
      { role: "user", text: value },
      { role: "agent", text: "Request added to the project plan. I will keep the next change focused and visible." },
    ]);
    setPrompt("");
  }

  return (
    <main className="flex h-screen min-h-0 flex-col overflow-hidden bg-forge-bg text-[#f1f3f5]">
      <TopBar name={name} email={email} initials={initials} onSignOut={signOut} onMenu={() => setSidebarOpen(!sidebarOpen)} onAssistant={() => setAssistantOpen(!assistantOpen)} />
      <div className="workspace-layout grid min-h-0 flex-1" style={{ "--workspace-columns": workspaceColumns } as CSSProperties}>
        {sidebarOpen && <WorkspaceSidebar active={activeNav} onChange={(label) => { setActiveNav(label); setSidebarOpen(false); }} onCollapse={() => setSidebarOpen(false)} />}
        <section className="flex min-w-0 flex-col bg-[#111419]">
          <WorkspaceHeader activeNav={activeNav} onSearch={() => setActiveNav("Files")} />
          <WorkspaceTabs active={activeTab} onChange={setActiveTab} />
          <div className="min-h-0 flex-1 overflow-auto p-4 lg:p-5">
            {activeTab === "preview" && <PreviewPanel device={device} setDevice={setDevice} />}
            {activeTab === "code" && <CodePanel />}
            {activeTab === "components" && <ComponentsPanel />}
            {activeTab === "console" && <ConsolePanel />}
          </div>
        </section>
        {assistantOpen && <AssistantPanel chat={messages} prompt={prompt} setPrompt={setPrompt} onSend={sendPrompt} onClose={() => setAssistantOpen(false)} />}
      </div>
      <MobileNav active={activeNav} onHome={() => setActiveNav("Overview")} onFiles={() => { setActiveNav("Files"); setSidebarOpen(true); }} onAgent={() => setAssistantOpen(true)} />
    </main>
  );
}
