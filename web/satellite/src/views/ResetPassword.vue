// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="reset-area" @keyup.enter="onResetClick">
        <div class="reset-area__logo-wrapper">
            <LogoIcon class="reset-area__logo-wrapper_logo" @click="onLogoClick" />
        </div>
        <div class="reset-area__content-area">
            <div class="reset-area__content-area__container" :class="{'success': isSuccessfulPasswordResetShown}">
                <template v-if="!isSuccessfulPasswordResetShown">
                    <h1 class="reset-area__content-area__container__title">Reset Password</h1>
                    <p class="reset-area__content-area__container__message">Please enter your new password.</p>
                    <div class="reset-area__content-area__container__input-wrapper password">
                        <HeaderlessInput
                            label="Password"
                            placeholder="Enter Password"
                            :error="passwordError"
                            width="100%"
                            height="46px"
                            is-password="true"
                            @setData="setPassword"
                            @showPasswordStrength="showPasswordStrength"
                            @hidePasswordStrength="hidePasswordStrength"
                        />
                        <PasswordStrength
                            :password-string="password"
                            :is-shown="isPasswordStrengthShown"
                        />
                    </div>
                    <div class="reset-area__content-area__container__input-wrapper">
                        <HeaderlessInput
                            label="Retype Password"
                            placeholder="Retype Password"
                            :error="repeatedPasswordError"
                            width="100%"
                            height="46px"
                            is-password="true"
                            @setData="setRepeatedPassword"
                        />
                    </div>
                    <p class="reset-area__content-area__container__button" @click.prevent="onResetClick">Reset Password</p>
                </template>
                <template v-else>
                    <KeyIcon />
                    <h2 class="reset-area__content-area__container__title success">Success!</h2>
                    <p class="reset-area__content-area__container__sub-title">
                        You have successfully changed your password.
                    </p>
                </template>
            </div>
            <router-link :to="loginPath" class="reset-area__content-area__login-link">
                Back to Login
            </router-link>
        </div>
    </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import HeaderlessInput from '@/components/common/HeaderlessInput.vue';
import PasswordStrength from '@/components/common/PasswordStrength.vue';

import LogoIcon from '@/../static/images/dcs-logo.svg';
import KeyIcon from '@/../static/images/resetPassword/success.svg';

import { AuthHttpApi } from '@/api/auth';
import { RouteConfig } from '@/router';
import { APP_STATE_ACTIONS } from '@/utils/constants/actionNames';
import { Validator } from '@/utils/validation';

@Component({
    components: {
        LogoIcon,
        HeaderlessInput,
        PasswordStrength,
        KeyIcon,
    },
})

export default class ResetPassword extends Vue {
    private token = '';
    private password = '';
    private repeatedPassword = '';

    private passwordError = '';
    private repeatedPasswordError = '';
    private isLoading = false;

    private readonly auth: AuthHttpApi = new AuthHttpApi();

    public isPasswordStrengthShown = false;

    public readonly loginPath: string = RouteConfig.Login.path;

    /**
     * Lifecycle hook on component destroy.
     * Sets view to default state.
     */
    public beforeDestroy(): void {
        if (this.isSuccessfulPasswordResetShown) {
            this.$store.dispatch(APP_STATE_ACTIONS.TOGGLE_SUCCESSFUL_PASSWORD_RESET);
        }
    }

    /**
     * Lifecycle hook after initial render.
     * Initializes recovery token from route param
     * and redirects to login if token doesn't exist.
     */
    public mounted(): void {
        if (this.$route.query.token) {
            this.token = this.$route.query.token.toString();
        } else {
            this.$router.push(RouteConfig.Login.path);
        }
    }

    /**
     * Returns whether the successful password reset area is shown.
     */
    public get isSuccessfulPasswordResetShown() : boolean {
        return this.$store.state.appStateModule.appState.isSuccessfulPasswordResetShown;
    }

    /**
     * Validates input fields and requests password reset.
     */
    public async onResetClick(): Promise<void> {
        if (this.isLoading) {
            return;
        }

        this.isLoading = true;

        if (!this.validateFields()) {
            this.isLoading = false;

            return;
        }

        try {
            await this.auth.resetPassword(this.token, this.password);
            this.$store.dispatch(APP_STATE_ACTIONS.TOGGLE_SUCCESSFUL_PASSWORD_RESET);
        } catch (error) {
            await this.$notify.error(error.message);
        }

        this.isLoading = false;
    }

    /**
     * Validates input values to satisfy expected rules.
     */
    private validateFields(): boolean {
        let isNoErrors = true;

        if (!Validator.password(this.password)) {
            this.passwordError = 'Invalid password';
            isNoErrors = false;
        }

        if (this.repeatedPassword !== this.password) {
            this.repeatedPasswordError = 'Password doesn\'t match';
            isNoErrors = false;
        }

        return isNoErrors;
    }

    /**
     * Makes password strength container visible.
     */
    public showPasswordStrength(): void {
        this.isPasswordStrengthShown = true;
    }

    /**
     * Hides password strength container.
     */
    public hidePasswordStrength(): void {
        this.isPasswordStrengthShown = false;
    }

    /**
     * Reloads the page.
     */
    public onLogoClick(): void {
        location.reload();
    }

    /**
     * Sets user's password field from value string.
     */
    public setPassword(value: string): void {
        this.password = value.trim();
        this.passwordError = '';
    }

    /**
     * Sets user's repeat password field from value string.
     */
    public setRepeatedPassword(value: string): void {
        this.repeatedPassword = value.trim();
        this.repeatedPasswordError = '';
    }
}
</script>

<style scoped lang="scss">
    .reset-area {
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: center;
        font-family: 'font_regular', sans-serif;
        background-color: #f5f6fa;
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        min-height: 100%;
        overflow-y: scroll;

        &__logo-wrapper {
            text-align: center;
            margin: 70px 0;

            &__logo {
                cursor: pointer;
            }
        }

        &__content-area {
            width: 100%;
            padding: 0 20px;
            margin-bottom: 50px;
            display: flex;
            flex-direction: column;
            align-items: center;
            box-sizing: border-box;

            &__container {
                width: 610px;
                padding: 60px 80px;
                display: flex;
                flex-direction: column;
                background-color: #fff;
                border-radius: 20px;
                box-sizing: border-box;

                &.success {
                    align-items: center;
                    text-align: center;
                }

                &__input-wrapper {
                    margin-top: 20px;

                    &.password {
                        position: relative;
                    }
                }

                &__title {
                    font-size: 24px;
                    margin: 10px 0;
                    letter-spacing: -0.100741px;
                    color: #252525;
                    font-family: 'font_bold', sans-serif;
                    font-weight: 800;

                    &.success {
                        font-size: 40px;
                        margin: 25px 0;
                    }
                }

                &__button {
                    font-family: 'font_regular', sans-serif;
                    font-weight: 700;
                    margin-top: 40px;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    background-color: #376fff;
                    border-radius: 50px;
                    color: #fff;
                    cursor: pointer;
                    width: 100%;
                    height: 48px;

                    &:hover {
                        background-color: #0059d0;
                    }
                }
            }

            &__login-link {
                font-family: 'font_medium', sans-serif;
                text-decoration: none;
                font-size: 14px;
                line-height: 18px;
                color: #376fff;
                margin-top: 50px;
            }
        }
    }

    @media screen and (max-width: 750px) {

        .reset-area {

            &__content-area {

                &__container {
                    width: 100%;
                }
            }
        }
    }

    @media screen and (max-width: 414px) {

        .reset-area {

            &__logo-wrapper {
                margin: 40px;
            }

            &__content-area {
                padding: 0;

                &__container {
                    padding: 60px 60px;
                    border-radius: 0;
                }
            }
        }
    }
</style>
