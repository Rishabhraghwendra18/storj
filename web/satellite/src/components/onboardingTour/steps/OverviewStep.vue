// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="overview-area">
        <div class="overview-area__header">
            <WelcomeLeft />
            <h1 class="overview-area__header__title">Welcome</h1>
            <WelcomeRight />
        </div>
        <p class="overview-area__subtitle">Let's get you started using Storj DCS</p>
        <p class="overview-area__question">Do you want to use web browser or command-line interface?</p>
        <div class="overview-area__routes">
            <OverviewContainer
                class="overview-area__routes__left-cont"
                is-web="true"
                title="Web browser"
                encryption="SERVER-SIDE ENCRYPTED"
                info="Start uploading files in the browser and instantly see how your data gets distributed over the Storj network around the world."
                encryption-container="By using the web browser you are opting in to server-side encryption."
                button-label="CONTINUE IN WEB"
                :on-click="onUploadInBrowserClick"
                :is-disabled="isLoading"
            />
            <OverviewContainer
                title="Command line"
                encryption="END-TO-END ENCRYPTED"
                info="The Uplink CLI is a command-line interface which allows you to upload and download files from the network, manage permissions and sharing."
                encryption-container="The Uplink CLI uses end-to-end encryption for object data, metadata and path data."
                button-label="CONTINUE IN CLI"
                :on-click="onUplinkCLIClick"
                :is-disabled="isLoading"
            />
        </div>
        <div
            class="overview-area__skip-button"
            :on-press="onSkipClick"
            @click.prevent="onSkipClick"
        >
            Skip and go directly to dashboard
        </div>
    </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import VButton from '@/components/common/VButton.vue';
import OverviewContainer from '@/components/onboardingTour/steps/common/OverviewContainer.vue';
import WelcomeLeft from '@/../static/images/onboardingTour/welcome-left.svg';
import WelcomeRight from '@/../static/images/onboardingTour/welcome-right.svg';

import { AnalyticsHttpApi } from '@/api/analytics';
import { RouteConfig } from '@/router';
import { PROJECTS_ACTIONS } from '@/store/modules/projects';
import { AnalyticsEvent } from '@/utils/constants/analyticsEventNames';

@Component({
    components: {
        VButton,
        OverviewContainer,
        WelcomeLeft,
        WelcomeRight
    },
})
export default class OverviewStep extends Vue {
    public isLoading = false;

    private readonly analytics: AnalyticsHttpApi = new AnalyticsHttpApi();

    /**
     * Lifecycle hook after initial render.
     * Sets area to needed state.
     */
    public mounted(): void {
        if (this.userHasProject) {
            this.$router.push(RouteConfig.OnboardingTour.with(RouteConfig.AccessGrant).path).catch(() => {return; });
        }
    }

    /**
     * Holds button click logic.
     * Creates untitled project and redirects to next step (creating access grant).
     */
    public async onUplinkCLIClick(): Promise<void> {
        if (this.isLoading) return;

        this.isLoading = true;

        await this.analytics.linkEventTriggered(AnalyticsEvent.PATH_SELECTED, 'CLI');

        try {
            await this.$store.dispatch(PROJECTS_ACTIONS.CREATE_DEFAULT_PROJECT);

            this.isLoading = false;

            await this.$router.push(RouteConfig.OnboardingTour.with(RouteConfig.AccessGrant).with(RouteConfig.AccessGrantName).path);
        } catch (error) {
            await this.$notify.error(error.message);
            this.isLoading = false;
        }
    }

    /**
     * Creates untitled project and redirects to objects page.
     */
    public async onUploadInBrowserClick(): Promise<void> {
        if (this.isLoading) return;

        this.isLoading = true;

        await this.analytics.linkEventTriggered(AnalyticsEvent.PATH_SELECTED, 'Continue in Browser');

        try {
            await this.$store.dispatch(PROJECTS_ACTIONS.CREATE_DEFAULT_PROJECT);

            this.isLoading = false;

            await this.$router.push(RouteConfig.Objects.path).catch(() => {return; });
        } catch (error) {
            await this.$notify.error(error.message);
            this.isLoading = false;
        }
    }

    /**
     * Holds button click logic.
     * Creates untitled project and redirects to project dashboard.
     */
    public async onSkipClick(): Promise<void> {
        if (this.isLoading) return;

        this.isLoading = true;

        try {
            await this.$store.dispatch(PROJECTS_ACTIONS.CREATE_DEFAULT_PROJECT);

            this.isLoading = false;

            await this.$router.push(RouteConfig.ProjectDashboard.path);
        } catch (error) {
            await this.$notify.error(error.message);
            this.isLoading = false;
        }
    }

    /**
     * Indicates if user has at least one project.
     */
    private get userHasProject(): boolean {
        return this.$store.state.projectsModule.projects.length > 0;
    }
}
</script>

<style scoped lang="scss">
    .overview-area {
        display: flex;
        flex-direction: column;
        align-items: center;
        font-family: 'font_regular', sans-serif;

        &__header {
            display: flex;

            &__title {
                font-family: 'font_bold', sans-serif;
                font-size: 48px;
                line-height: 48px;
                letter-spacing: 1px;
                color: #14142b;
                margin: 0 20px 20px 20px;
            }
        }

        &__subtitle,
        &__question {
            font-family: 'font_medium', sans-serif;
            font-size: 16px;
            line-height: 32px;
            text-align: center;
            letter-spacing: 0.75px;
            color: #14142a;
        }

        &__question {
            font-family: 'font_regular', sans-serif;
        }

        &__routes {
            margin-top: 70px;
            display: flex;
            align-items: center;

            &__left-cont {
                margin-right: 45px;
            }
        }

        &__skip-button {
            margin: 50px 0 40px 0;
            color: #b7c1ca;
            cursor: pointer;
            text-decoration: underline !important;

            &:hover {
                text-decoration: underline;
            }
        }
    }
</style>
