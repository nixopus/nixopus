"use client"
import { LoginForm } from "@/components/login-form"
import useLogin from "./hooks/use-login"
import { ErrorBoundary } from "@/components/ui/error-handler"
import { useAppSelector } from "@/redux/hooks"
import { useRouter } from "next/navigation"

export default function LoginPage() {
  const { email, password, handleEmailChange, handlePasswordChange, handleLogin, isLoading, error } = useLogin()
  const authenticated = useAppSelector(state => state.auth.isAuthenticated)
  const router = useRouter()

  if (authenticated) {
    router.push("/")
    return < ></>
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <LoginForm
          email={email}
          password={password}
          handleEmailChange={handleEmailChange}
          handlePasswordChange={handlePasswordChange}
          handleLogin={handleLogin}
          isLoading={isLoading}
        />
      </div>
      {
        error && (
          <ErrorBoundary
            errors={[
              { error, title: "Login Error" }]}
          />
        )
      }
    </div>
  )
}
