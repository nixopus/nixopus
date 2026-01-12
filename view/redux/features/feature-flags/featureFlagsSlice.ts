import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { FeatureFlag } from '@/packages/types/feature-flags';

interface FeatureFlagsState {
  features: FeatureFlag[];
  isLoading: boolean;
  error: string | null;
}

const initialState: FeatureFlagsState = {
  features: [],
  isLoading: false,
  error: null
};

const featureFlagsSlice = createSlice({
  name: 'featureFlags',
  initialState,
  reducers: {
    setFeatures: (state, action: PayloadAction<FeatureFlag[]>) => {
      state.features = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    updateFeature: (state, action: PayloadAction<{ featureName: string; isEnabled: boolean }>) => {
      const feature = state.features.find((f) => f.feature_name === action.payload.featureName);
      if (feature) {
        feature.is_enabled = action.payload.isEnabled;
      }
    }
  }
});

export const { setFeatures, setLoading, setError, updateFeature } = featureFlagsSlice.actions;
export default featureFlagsSlice.reducer;
