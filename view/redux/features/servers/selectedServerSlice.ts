import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface SelectedServerState {
  selectedServerId: string | null;
}

const initialState: SelectedServerState = {
  selectedServerId: null
};

export const selectedServerSlice = createSlice({
  name: 'selectedServer',
  initialState,
  reducers: {
    setSelectedServer: (state, action: PayloadAction<string | null>) => {
      state.selectedServerId = action.payload;
    }
  }
});

export const { setSelectedServer } = selectedServerSlice.actions;

export const selectSelectedServerId = (state: { selectedServer: SelectedServerState }) =>
  state.selectedServer.selectedServerId;

export default selectedServerSlice.reducer;
