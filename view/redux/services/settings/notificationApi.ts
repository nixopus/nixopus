import { USER_NOTIFICATION_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  CreateSMTPConfigRequest,
  GetPreferencesResponse,
  SMTPConfig,
  UpdatePreferenceRequest,
  UpdateSMTPConfigRequest,
  WebhookConfig,
  CreateWebhookConfigRequest,
  UpdateWebhookConfigRequest,
  DeleteWebhookConfigRequest,
  GetWebhookConfigRequest
} from '@/redux/types/notification';

export const notificationApi = createApi({
  reducerPath: 'notificationApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Notification'],
  endpoints: (builder) => ({
    getSMTPConfigurations: builder.query<SMTPConfig, string>({
      query: (organizationId) => ({
        url: USER_NOTIFICATION_SETTINGS.GET_SMTP,
        method: 'GET',
        params: { id: organizationId }
      }),
      providesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: SMTPConfig }) => {
        return response.data;
      }
    }),
    createSMPTConfiguration: builder.mutation<SMTPConfig, CreateSMTPConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.ADD_SMTP,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: SMTPConfig }) => {
        return response.data;
      }
    }),
    updateSMTPConfiguration: builder.mutation<SMTPConfig, UpdateSMTPConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.UPDATE_SMTP,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: SMTPConfig }) => {
        return response.data;
      }
    }),
    deleteSMTPConfiguration: builder.mutation<void, { id: string; organization_id: string }>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.DELETE_SMTP,
        method: 'DELETE',
        params: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }]
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
    updateNotificationPreferences: builder.mutation<void, UpdatePreferenceRequest>({
      query: (payload) => ({
        url: USER_NOTIFICATION_SETTINGS.UPDATE_PREFERENCES,
        method: 'POST',
        body: payload
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }]
    }),
    getWebhookConfig: builder.query<WebhookConfig, GetWebhookConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.GET_WEBHOOK + `/${data.type}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: WebhookConfig }) => {
        return response.data;
      }
    }),
    createWebhookConfig: builder.mutation<WebhookConfig, CreateWebhookConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.CREATE_WEBHOOK,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: WebhookConfig }) => {
        return response.data;
      }
    }),
    updateWebhookConfig: builder.mutation<WebhookConfig, UpdateWebhookConfigRequest>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.UPDATE_WEBHOOK,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
      transformResponse: (response: { data: WebhookConfig }) => {
        return response.data;
      }
    }),
    deleteWebhookConfig: builder.mutation<void, { type: string; organization_id: string }>({
      query: (data) => ({
        url: USER_NOTIFICATION_SETTINGS.DELETE_WEBHOOK,
        method: 'DELETE',
        params: data
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }]
    })
  })
});

export const {
  useGetSMTPConfigurationsQuery,
  useCreateSMPTConfigurationMutation,
  useUpdateSMTPConfigurationMutation,
  useDeleteSMTPConfigurationMutation,
  useGetNotificationPreferencesQuery,
  useUpdateNotificationPreferencesMutation,
  useGetWebhookConfigQuery,
  useCreateWebhookConfigMutation,
  useUpdateWebhookConfigMutation,
  useDeleteWebhookConfigMutation
} = notificationApi;
