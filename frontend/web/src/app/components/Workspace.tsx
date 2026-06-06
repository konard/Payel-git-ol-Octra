import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';
import { Canvas } from './Canvas';
import { Chat, type ChatMessage } from './Chat';
import { SolutionViewer } from './SolutionViewer';
import { BottomInput, type TaskData } from './BottomInput';
import { Sidebar } from '../../components/Sidebar';
import { DesktopFileExplorer } from '../../desktop/DesktopFileExplorer';
import { isDesktopApp } from '../../desktop/bridge';

interface WorkspaceProps {
  // The Canvas needs a mode prop to decide whether to show its "add agent"
  // affordance; on desktop the centre always renders the canvas graph, so we
  // pass 'canvas' through. onModeChange is forwarded for the same reason.
  onModeChange: (mode: 'canvas' | 'chat' | 'solution') => void;
  hasUnreadMessages: boolean;
  chatMessages: ChatMessage[];
  onMarkAsRead: (messageId: string) => void;
  onCreateTask: (data: TaskData) => void;
  onStopTask: () => void;
  isSubmitting: boolean;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onSelectChat: (chatId: string) => void;
  onNewChat: (chatId?: string) => void;
  // Whether the left Sessions dock is visible. It is toggled from the header so
  // the centre and Solution panes can take the full width. The Solution dock is
  // always visible on desktop (the redundant toggles were removed per issue #40).
  sessionsOpen: boolean;
}

// A reusable rounded "window card" wrapper that gives every pane the framed look
// from the reference design while preserving the existing surface colours.
function PaneCard({ children, dataTour }: { children: React.ReactNode; dataTour?: string }) {
  return (
    <div
      className="h-full min-h-0 overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] shadow-sm"
      data-tour={dataTour}
    >
      <div className="flex h-full min-h-0 flex-col">{children}</div>
    </div>
  );
}

// A slim grabbable divider between resizable panes. Orientation follows the
// parent PanelGroup direction: a vertical bar for horizontal groups, a
// horizontal bar for vertical (stacked) groups.
function ResizeHandle({ orientation = 'vertical' }: { orientation?: 'vertical' | 'horizontal' }) {
  if (orientation === 'horizontal') {
    return (
      <PanelResizeHandle className="group relative h-2 shrink-0">
        <div className="absolute inset-x-0 top-1/2 h-px -translate-y-1/2 bg-[var(--border)] transition-colors group-hover:bg-[var(--accent)] group-data-[resize-handle-state=drag]:bg-[var(--accent)]" />
      </PanelResizeHandle>
    );
  }
  return (
    <PanelResizeHandle className="group relative w-2 shrink-0">
      <div className="absolute inset-y-0 left-1/2 w-px -translate-x-1/2 bg-[var(--border)] transition-colors group-hover:bg-[var(--accent)] group-data-[resize-handle-state=drag]:bg-[var(--accent)]" />
    </PanelResizeHandle>
  );
}

export function Workspace({
  onModeChange,
  hasUnreadMessages,
  chatMessages,
  onMarkAsRead,
  onCreateTask,
  onStopTask,
  isSubmitting,
  isExpanded,
  onToggleExpand,
  onSelectChat,
  onNewChat,
  sessionsOpen,
}: WorkspaceProps) {
  // The centre column stacks the Canvas (top) over the Octra Boss chat (bottom)
  // as two resizable panes, then shares a SINGLE input docked at the bottom of
  // the column. The owner asked for one unified field instead of a separate
  // input for the canvas and another for the chat (issue #40 feedback). Because
  // both panes are always visible, the old Canvas/Chat/Solution header toggles
  // were redundant and were removed too.
  const center = (
    <PaneCard dataTour="workflow-workspace">
      <PanelGroup direction="vertical" className="min-h-0 flex-1">
        <Panel id="center-canvas" order={1} defaultSize={58} minSize={25}>
          {/* The Canvas root is `flex-1`, so its wrapper must be a flex column —
              otherwise the ReactFlow surface collapses to height 0 and the dotted
              grid / zoom controls never render (issue #40 feedback). */}
          <div className="flex h-full min-h-0 flex-col">
            <Canvas mode="canvas" onModeChange={onModeChange} hasUnreadMessages={hasUnreadMessages} />
          </div>
        </Panel>

        <ResizeHandle orientation="horizontal" />

        <Panel id="center-chat" order={2} defaultSize={42} minSize={20}>
          <div className="h-full min-h-0">
            <Chat messages={chatMessages} onMarkAsRead={onMarkAsRead} />
          </div>
        </Panel>
      </PanelGroup>

      <BottomInput
        onSubmit={onCreateTask}
        onStop={onStopTask}
        isSubmitting={isSubmitting}
        isExpanded={isExpanded}
        onToggleExpand={onToggleExpand}
      />
    </PaneCard>
  );

  // Re-key the group when the Sessions dock visibility changes so
  // react-resizable-panels lays out cleanly instead of trying to reconcile a
  // changed panel set.
  // The desktop app adds a filesystem-backed Explorer dock on the far left. It is
  // present only inside Electron, so the web layout is unchanged. Include it in
  // the layout key so the panel group reconciles cleanly when it is present.
  const showExplorer = isDesktopApp();
  const layoutKey =
    (sessionsOpen ? 'sessions-open' : 'sessions-closed') + (showExplorer ? '-explorer' : '');

  return (
    <div className="relative min-h-0 flex-1 bg-[var(--background)] p-2">
      {/* Files opened from the desktop Explorer render in the Solution files
          panel (SolutionViewer), not a separate viewer window (issue #50). */}
      <PanelGroup key={layoutKey} direction="horizontal" className="h-full">
        {showExplorer && (
          <>
            <Panel id="explorer" order={0} defaultSize={20} minSize={14} maxSize={36}>
              <div className="h-full min-h-0 overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] shadow-sm">
                <DesktopFileExplorer />
              </div>
            </Panel>
            <ResizeHandle />
          </>
        )}
        {sessionsOpen && (
          <>
            <Panel id="sessions" order={1} defaultSize={20} minSize={14} maxSize={32}>
              <PaneCard dataTour="chat-sessions">
                <Sidebar
                  variant="dock"
                  isOpen
                  onClose={() => {}}
                  onSelectChat={onSelectChat}
                  onNewChat={onNewChat}
                />
              </PaneCard>
            </Panel>
            <ResizeHandle />
          </>
        )}

        <Panel id="center" order={2} minSize={30}>
          {center}
        </Panel>

        <ResizeHandle />
        <Panel id="solution" order={3} defaultSize={32} minSize={20} maxSize={50}>
          <PaneCard dataTour="solution-pane">
            <SolutionViewer />
          </PaneCard>
        </Panel>
      </PanelGroup>
    </div>
  );
}
