import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';
import { Canvas } from './Canvas';
import { Chat, type ChatMessage } from './Chat';
import { SolutionViewer } from './SolutionViewer';
import { BottomInput, type TaskData } from './BottomInput';
import { ChatInput } from './ChatInput';
import { Sidebar } from '../../components/Sidebar';

interface WorkspaceProps {
  // The center pane mirrors the existing single-pane modes. On desktop we only
  // ever surface 'canvas' or 'chat' in the centre because the solution lives in
  // its own docked pane, but we keep the full union so callers can stay generic.
  mode: 'canvas' | 'chat' | 'solution';
  onModeChange: (mode: 'canvas' | 'chat' | 'solution') => void;
  hasUnreadMessages: boolean;
  chatMessages: ChatMessage[];
  onMarkAsRead: (messageId: string) => void;
  onCreateTask: (data: TaskData) => void;
  onStopTask: () => void;
  isSubmitting: boolean;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onSendChatMessage: (message: string) => void;
  onSelectChat: (chatId: string) => void;
  onNewChat: (chatId?: string) => void;
  // Whether the left Sessions dock and right Solution dock are visible. They are
  // toggled from the header so the centre canvas can take the full width.
  sessionsOpen: boolean;
  solutionOpen: boolean;
}

// A reusable rounded "window card" wrapper that gives every pane the framed look
// from the reference design while preserving the existing surface colours.
function PaneCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-full min-h-0 overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] shadow-sm">
      <div className="flex h-full min-h-0 flex-col">{children}</div>
    </div>
  );
}

// A slim grabbable divider between resizable panes.
function ResizeHandle() {
  return (
    <PanelResizeHandle className="group relative w-2 shrink-0">
      <div className="absolute inset-y-0 left-1/2 w-px -translate-x-1/2 bg-[var(--border)] transition-colors group-hover:bg-[var(--accent)] group-data-[resize-handle-state=drag]:bg-[var(--accent)]" />
    </PanelResizeHandle>
  );
}

export function Workspace({
  mode,
  onModeChange,
  hasUnreadMessages,
  chatMessages,
  onMarkAsRead,
  onCreateTask,
  onStopTask,
  isSubmitting,
  isExpanded,
  onToggleExpand,
  onSendChatMessage,
  onSelectChat,
  onNewChat,
  sessionsOpen,
  solutionOpen,
}: WorkspaceProps) {
  // On desktop the solution has a dedicated dock, so the centre only ever shows
  // the canvas or the chat. Treat the 'solution' mode as 'canvas' here.
  const centerMode: 'canvas' | 'chat' = mode === 'chat' ? 'chat' : 'canvas';

  const center = (
    <PaneCard>
      <div className="min-h-0 flex-1">
        {centerMode === 'canvas' ? (
          <Canvas mode="canvas" onModeChange={onModeChange} hasUnreadMessages={hasUnreadMessages} />
        ) : (
          <Chat messages={chatMessages} onMarkAsRead={onMarkAsRead} />
        )}
      </div>
      {centerMode === 'canvas' ? (
        <BottomInput
          onSubmit={onCreateTask}
          onStop={onStopTask}
          isSubmitting={isSubmitting}
          isExpanded={isExpanded}
          onToggleExpand={onToggleExpand}
        />
      ) : (
        <ChatInput onSendMessage={onSendChatMessage} />
      )}
    </PaneCard>
  );

  // Re-key the group when the visible docks change so react-resizable-panels lays
  // out cleanly instead of trying to reconcile a changed panel set.
  const layoutKey = `${sessionsOpen ? 's' : ''}-${solutionOpen ? 'v' : ''}`;

  return (
    <div className="min-h-0 flex-1 bg-[var(--background)] p-2">
      <PanelGroup key={layoutKey} direction="horizontal" className="h-full">
        {sessionsOpen && (
          <>
            <Panel id="sessions" order={1} defaultSize={20} minSize={14} maxSize={32}>
              <PaneCard>
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

        {solutionOpen && (
          <>
            <ResizeHandle />
            <Panel id="solution" order={3} defaultSize={32} minSize={20} maxSize={50}>
              <PaneCard>
                <SolutionViewer />
              </PaneCard>
            </Panel>
          </>
        )}
      </PanelGroup>
    </div>
  );
}
