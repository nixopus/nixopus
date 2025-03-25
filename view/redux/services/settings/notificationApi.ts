import { USER_NOTIFICATION_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  CreateSMTPConfigRequest,
  GetPreferencesResponse,
  SMTPConfig,
  UpdatePreferenceRequest,
  UpdateSMTPConfigRequest
} from '@/redux/types/notification';

export const notificationApi = createApi({
  reducerPath: 'notificationApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Notification'],
  endpoints: (builder) => ({
    getSMTPConfigurations: builder.query<SMTPConfig, string>({
      query: (organizationId) => ({
        url: USER_NOTIFICATION_SETTINGS.GET_SMTP + `?id=${organizationId}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: SMTPConfig }) => {
        return response.data;
      }
    }),
    createSMPTConfiguration: builder.mutation<null, CreateSMTPConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.ADD_SMTP,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    updateSMTPConfiguration: builder.mutation<null, UpdateSMTPConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.UPDATE_SMTP,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    deleteSMTPConfiguration: builder.mutation<null, string>({
      query: (id) => ({
        url: USER_NOTIFICATION_SETTINGS.DELETE_SMTP,
        method: 'DELETE',
        params: { id }
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    getNotificationPreferences: builder.query<GetPreferencesResponse, void>({
      query: () => ({
        url: USER_NOTIFICATION_SETTINGS.GET_PREFERENCES,
        method: 'GET'
      }),
      providesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: GetPreferencesResponse }) => {
        return response.data;
      }
    }),
    updateNotificationPreferences: builder.mutation<null, UpdatePreferenceRequest>({
      query: (payload) => ({
        url: USER_NOTIFICATION_SETTINGS.UPDATE_PREFERENCES,
        method: 'POST',
        body: payload
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetSMTPConfigurationsQuery,
  useCreateSMPTConfigurationMutation,
  useUpdateSMTPConfigurationMutation,
  useDeleteSMTPConfigurationMutation,
  useGetNotificationPreferencesQuery,
  useUpdateNotificationPreferencesMutation
} = notificationApi;
