import { combineReducers, configureStore, ConfigureStoreOptions } from '@reduxjs/toolkit';
import {
  persistStore,
  persistReducer,
  FLUSH,
  REHYDRATE,
  PAUSE,
  PERSIST,
  PURGE,
  REGISTER
} from 'redux-persist';
import createWebStorage from 'redux-persist/lib/storage/createWebStorage';
import { authApi } from '@/redux/services/users/authApi';
import authReducer from '@/redux/features/users/authSlice';
import { userApi } from '@/redux/services/users/userApi';
import userSlice from '@/redux/features/users/userSlice';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import githubConnector from './features/github-connector/githubConnectorSlice';
import { deployApi } from './services/deploy/applicationsApi';
import { healthcheckApi } from './services/deploy/healthcheckApi';
import { fileManagersApi } from './services/file-manager/fileManagersApi';
import { auditApi } from './services/audit';
import { FeatureFlagsApi } from './services/feature-flags/featureFlagsApi';
import { containerApi } from './services/container/containerApi';
import { imagesApi } from './services/container/imagesApi';
import { extensionsApi } from './services/extensions/extensionsApi';
const createNoopStorage = () => ({
  getItem: (_key: string) => Promise.resolve(null),
  setItem: (_key: string, value: any) => Promise.resolve(value),
  removeItem: (_key: string) => Promise.resolve()
});

const storage = typeof window !== 'undefined' ? createWebStorage('local') : createNoopStorage();

const persistConfig = {
  key: 'root',
  version: 1,
  storage,
  whitelist: ['auth', 'user'],
  migrate: (state: any) => {
    if (!state) return Promise.resolve(undefined);
    return Promise.resolve(state);
  }
};

const rootReducer = combineReducers({
  [authApi.reducerPath]: authApi.reducer,
  auth: authReducer,
  [userApi.reducerPath]: userApi.reducer,
  notificationApi: notificationApi.reducer,
  [domainsApi.reducerPath]: domainsApi.reducer,
  [GithubConnectorApi.reducerPath]: GithubConnectorApi.reducer,
  githubConnector: githubConnector,
  [deployApi.reducerPath]: deployApi.reducer,
  [healthcheckApi.reducerPath]: healthcheckApi.reducer,
  user: userSlice,
  fileManagersApi: fileManagersApi.reducer,
  [auditApi.reducerPath]: auditApi.reducer,
  [FeatureFlagsApi.reducerPath]: FeatureFlagsApi.reducer,
  [containerApi.reducerPath]: containerApi.reducer,
  [imagesApi.reducerPath]: imagesApi.reducer,
  [extensionsApi.reducerPath]: extensionsApi.reducer
});

type RootReducer = ReturnType<typeof rootReducer>;

const appReducer = (state: RootReducer | undefined, action: { type: string }) => {
  if (action.type === 'RESET_STATE') {
    return rootReducer(undefined, action);
  }
  return rootReducer(state, action);
};

const persistedReducer = persistReducer(persistConfig, appReducer);

const storeOptions: ConfigureStoreOptions = {
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER]
      },
      immutableCheck: process.env.NODE_ENV === 'development',
      thunk: true
    }).concat([
      authApi.middleware,
      userApi.middleware,
      notificationApi.middleware,
      domainsApi.middleware,
      GithubConnectorApi.middleware,
      deployApi.middleware,
      healthcheckApi.middleware,
      fileManagersApi.middleware,
      auditApi.middleware,
      FeatureFlagsApi.middleware,
      containerApi.middleware,
      imagesApi.middleware,
      extensionsApi.middleware
    ]),
  devTools: process.env.NODE_ENV === 'development'
};

export const store = configureStore(storeOptions);

export const persistor = persistStore(store);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
