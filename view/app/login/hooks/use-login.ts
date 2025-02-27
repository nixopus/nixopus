import { useLoginUserMutation } from '@/redux/services/users/authApi'
import React from 'react'

function useLogin() {
    const [email, setEmail] = React.useState('')
    const [password, setPassword] = React.useState('')
    const [loginUser, { isLoading, error }] = useLoginUserMutation()

    const handleEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setEmail(event.target.value)
    }

    const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setPassword(event.target.value)
    }

    const handleLogin = async () => {
        await loginUser({ email, password })
    }

    return {
        email,
        password,
        handleEmailChange,
        handlePasswordChange,
        handleLogin,
        isLoading,
        error
    }
}

export default useLogin