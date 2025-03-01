'use client';

import React from 'react';
import { Lock, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TabsContent } from '@/components/ui/tabs';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface SecuritySectionProps {
  emailSent: boolean;
  isLoading: boolean;
  handlePasswordResetRequest: () => void;
}

function SecuritySection({
  emailSent,
  isLoading,
  handlePasswordResetRequest
}: SecuritySectionProps) {
  return (
    <TabsContent value="security" className="space-y-4 mt-4">
      <Card>
        <CardHeader>
          <CardTitle>Password</CardTitle>
          <CardDescription>Change your password </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label className="flex items-center gap-2">
              <Lock size={16} />
              Reset Password
            </Label>
            <p className="text-sm text-muted-foreground">
              We'll send a password reset link to your email address
            </p>
          </div>

          {emailSent ? (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertTitle>Email Sent!</AlertTitle>
              <AlertDescription>Check your inbox for a password reset link</AlertDescription>
            </Alert>
          ) : (
            <Button
              onClick={handlePasswordResetRequest}
              disabled={isLoading}
              variant="outline"
              className="w-full lg:w-auto"
            >
              Send Password Reset Link
            </Button>
          )}
        </CardContent>
      </Card>
    </TabsContent>
  );
}

export default SecuritySection;
