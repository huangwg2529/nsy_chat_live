import { ArrowDownToLine, CalendarDays, Info, Search, SendIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import useChatStore from "@/stores/chat-store";
import AboutDialog from "../about/AboutDialog";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import ChatList from "./ChatList";
import SearchBar from "./SearchBar";
import Watermark from "./Watermark";

const ChatPage = () => {
  const rooms = useChatStore((s) => s.rooms);
  const selectedRoom = useChatStore((s) => s.selectedRoom);
  const isLoadingRooms = useChatStore((s) => s.isLoadingRooms);
  const isLoadingMessages = useChatStore((s) => s.isLoadingMessages);
  const error = useChatStore((s) => s.error);
  const loadRooms = useChatStore((s) => s.loadRooms);
  const selectRoom = useChatStore((s) => s.selectRoom);
  const jumpToDate = useChatStore((s) => s.jumpToDate);
  const requestScrollToBottom = useChatStore((s) => s.requestScrollToBottom);

  // 搜索相关状态
  const isSearchMode = useChatStore((s) => s.searchQuery.length > 0);
  const clearSearch = useChatStore((s) => s.clearSearch);

  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isAboutDialogOpen, setIsAboutDialogOpen] = useState(false);
  const [dateValue, setDateValue] = useState("");
  const dateInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    void loadRooms();
  }, [loadRooms]);

  // 快捷键监听
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "f") {
        e.preventDefault();
        setIsSearchOpen(true);
      }
      if (e.key === "Escape") {
        if (isSearchOpen) setIsSearchOpen(false);
        if (isSearchMode) clearSearch();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isSearchOpen, isSearchMode, clearSearch]);

  const selectedRoomKey = selectedRoom
    ? roomOptionValue(selectedRoom.chatRoomId, selectedRoom.displayName)
    : "";
  const isInitialLoading = isLoadingRooms || (isLoadingMessages && !selectedRoom);

  const handleRoomChange = (value: string) => {
    const nextRoom = rooms.find(
      (room) => roomOptionValue(room.chatRoomId, room.displayName) === value,
    );
    if (nextRoom) {
      void selectRoom(nextRoom);
    }
  };

  if (isInitialLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-muted/60">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-muted/60 md:mx-56 md:border-l md:border-r md:border-muted/20 relative">
      <Watermark text={selectedRoom?.displayName || "replive_web"} />

      {/* Header Area */}
      <header className="flex items-center justify-between px-4 py-3 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          <Avatar className="h-8 w-8 shrink-0">
            <AvatarImage
              src={selectedRoom?.avatarUrl}
              alt={selectedRoom?.displayName}
            />
            <AvatarFallback>
              {selectedRoom?.displayName?.slice(0, 1) || "R"}
            </AvatarFallback>
          </Avatar>
          <select
            value={selectedRoomKey}
            onChange={(event) => handleRoomChange(event.target.value)}
            disabled={rooms.length === 0 || isLoadingRooms}
            className="min-w-0 max-w-[180px] rounded-md border border-transparent bg-transparent px-2 py-1 text-sm font-semibold outline-none transition-colors hover:bg-accent focus:border-border focus:bg-background md:max-w-[260px]"
            title="切换聊天对象"
          >
            {rooms.length === 0 ? (
              <option value="">暂无聊天对象</option>
            ) : (
              rooms.map((room) => (
                <option
                  key={roomOptionValue(room.chatRoomId, room.displayName)}
                  value={roomOptionValue(room.chatRoomId, room.displayName)}
                >
                  {room.displayName}
                </option>
              ))
            )}
          </select>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-8 gap-1 px-2 text-xs text-muted-foreground"
              title="按日期跳转"
              disabled={!selectedRoom}
              onClick={() => {
                const input = dateInputRef.current;
                if (!input) return;
                if (typeof input.showPicker === "function") {
                  input.showPicker();
                  return;
                }
                input.focus();
                input.click();
              }}
            >
              <CalendarDays className="h-4 w-4" />
              <span className="hidden md:inline">按日期跳转</span>
              {dateValue && <span>{dateValue}</span>}
            </Button>
            <input
              ref={dateInputRef}
              type="date"
              value={dateValue}
              onChange={(event) => {
                const nextDate = event.target.value;
                setDateValue(nextDate);
                if (nextDate) {
                  void jumpToDate(nextDate);
                }
              }}
              className="pointer-events-none absolute left-0 top-0 h-px w-px opacity-0"
              disabled={!selectedRoom}
              tabIndex={-1}
            />
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => void requestScrollToBottom()}
            disabled={!selectedRoom}
            className="h-8 w-8"
            title="拉取最新消息并滑到底部"
          >
            <ArrowDownToLine className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              // 切换搜索栏显隐
              setIsSearchOpen((prev) => !prev);
              // 如果关闭搜索栏，顺便清空搜索状态
              if (isSearchOpen) clearSearch();
            }}
            className={`h-8 px-2 ${isSearchOpen || isSearchMode ? "text-primary bg-primary/10" : ""}`}
            title="按关键字搜索"
          >
            <Search className="h-4 w-4" />
            <span className="hidden text-xs md:inline">按关键字搜索</span>
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setIsAboutDialogOpen(true)}
            className="h-8 w-8"
          >
            <Info className="h-4 w-4" />
          </Button>
        </div>
      </header>

      {error && (
        <div className="border-b bg-destructive/10 px-4 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Search Bar Overlay */}
      <div
        className={`transition-all duration-300 ease-in-out overflow-hidden ${isSearchOpen ? "max-h-16 opacity-100" : "max-h-0 opacity-0"}`}
      >
        <SearchBar
          isOpen={isSearchOpen}
          onClose={() => {
            setIsSearchOpen(false);
            clearSearch();
          }}
        />
      </div>

      <AboutDialog
        isOpen={isAboutDialogOpen}
        onClose={() => setIsAboutDialogOpen(false)}
      />

      {/* 核心列表区域 */}
      <ChatList />

      {/* Footer Input Area */}
      <footer className="border-t bg-background px-4 py-3 sticky bottom-0 z-50 safe-area-bottom">
        <div className="flex items-center gap-2">
          <Input
            placeholder="不支持发送消息"
            disabled
            className="flex-1 bg-muted/50"
          />
          <Button disabled size="icon" className="shrink-0">
            <SendIcon className="h-4 w-4" />
          </Button>
        </div>
      </footer>
    </div>
  );
};

function roomOptionValue(chatRoomId: string, displayName: string) {
  return chatRoomId || displayName;
}

export default ChatPage;
