import { NotificationService } from "./types";

const notificationService: NotificationService = {
    saveEmailConfig: async (config: Record<string, string>) => {
        console.log('Saving email config:', config);
        return { success: true };
    },

    saveWebhookConfig: async (platform: string, config: Record<string, string>) => {
        console.log(`Saving ${platform} webhook config:`, config);
        return { success: true };
    },

    testEmailConnection: async (config: Record<string, string>) => {
        console.log('Testing email connection with config:', config);
        await new Promise(resolve => setTimeout(resolve, 1500));
        return { success: true, message: 'Test email sent successfully!' };
    },

    testWebhookConnection: async (platform: string, config: Record<string, string>) => {
        console.log(`Testing ${platform} webhook with config:`, config);
        await new Promise(resolve => setTimeout(resolve, 1500));
        return { success: true, message: `Test ${platform} notification sent successfully!` };
    },

    saveNotificationPreferences: async (preferences: Record<string, boolean>) => {
        console.log('Saving notification preferences:', preferences);
        return { success: true };
    }
};

export default notificationService;