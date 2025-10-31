import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { setActiveServer, setActiveServerId, clearActiveServer } from '@/redux/features/server/serverSlice';
import { Server } from '@/redux/types/server';

export function useActiveServer() {
  const dispatch = useAppDispatch();
  const activeServer = useAppSelector((state) => state.server.activeServer);
  const activeServerId = useAppSelector((state) => state.server.activeServerId);

  const setServer = (server: Server | null) => {
    dispatch(setActiveServer(server));
  };

  const setServerId = (serverId: string | null) => {
    dispatch(setActiveServerId(serverId));
  };

  const clearServer = () => {
    dispatch(clearActiveServer());
  };

  return {
    activeServer,
    activeServerId,
    setServer,
    setServerId,
    clearServer,
    isDefaultSelected: activeServerId === null
  };
}
