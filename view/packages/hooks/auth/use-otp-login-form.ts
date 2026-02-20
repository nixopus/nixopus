import { useState, useEffect, useRef } from 'react';
import { z } from 'zod';

export interface UseOtpLoginFormProps {
  email: string;
  otp: string;
  handleSendOtp: () => void;
  handleVerifyOtp: () => void;
  isVerifyingOtp?: boolean;
}

export function useOtpLoginForm({
  email,
  otp,
  handleSendOtp,
  handleVerifyOtp,
  isVerifyingOtp = false
}: UseOtpLoginFormProps) {
  const [emailError, setEmailError] = useState('');
  const [otpError, setOtpError] = useState('');
  const hasAutoSubmittedRef = useRef(false);

  const emailSchema = z.object({
    email: z.string().min(1, 'Email is required').email('Please enter a valid Email')
  });

  const otpSchema = z.object({
    otp: z.string().min(1, 'OTP is required').length(6, 'OTP must be 6 digits')
  });

  // Auto-submit when OTP is complete (6 digits)
  useEffect(() => {
    if (otp.length === 6 && !isVerifyingOtp && !hasAutoSubmittedRef.current) {
      hasAutoSubmittedRef.current = true;
      handleVerifyOtp();
    }
    // Reset auto-submit flag when OTP is cleared or incomplete
    if (otp.length < 6) {
      hasAutoSubmittedRef.current = false;
    }
  }, [otp, isVerifyingOtp, handleVerifyOtp]);

  const handleSendOtpClick = (): void => {
    setEmailError('');

    const result = emailSchema.safeParse({
      email: email ?? ''
    });

    if (!result.success) {
      const fieldErrors = result.error.flatten().fieldErrors;
      if (fieldErrors.email && fieldErrors.email[0]) {
        setEmailError(fieldErrors.email[0]);
      }
      return;
    }

    handleSendOtp();
  };

  const handleVerifyOtpClick = (): void => {
    setOtpError('');

    const result = otpSchema.safeParse({
      otp: otp ?? ''
    });

    if (!result.success) {
      const fieldErrors = result.error.flatten().fieldErrors;
      if (fieldErrors.otp && fieldErrors.otp[0]) {
        setOtpError(fieldErrors.otp[0]);
      }
      return;
    }

    handleVerifyOtp();
  };

  return {
    emailError,
    otpError,
    handleSendOtpClick,
    handleVerifyOtpClick
  };
}
