// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="buckets-dropdown">
        <div class="buckets-dropdown__container">
            <p class="buckets-dropdown__container__all" @click.stop="selectAllBuckets">
                All
            </p>
            <label class="buckets-dropdown__container__search">
                <input
                    v-model="bucketSearch"
                    class="buckets-dropdown__container__search__input"
                    placeholder="Search buckets"
                    type="text"
                >
            </label>
            <div
                v-for="(name, index) in bucketsList"
                :key="index"
                class="buckets-dropdown__container__choices"
            >
                <div
                    class="buckets-dropdown__container__choices__item"
                    :class="{ selected: isNameSelected(name) }"
                    @click.stop="toggleBucketSelection(name)"
                >
                    <div class="buckets-dropdown__container__choices__item__left">
                        <SelectionIcon class="buckets-dropdown__container__choices__item__left__icon" />
                        <p class="buckets-dropdown__container__choices__item__left__label">{{ name }}</p>
                    </div>
                    <UnselectIcon
                        v-if="isNameSelected(name)"
                        class="buckets-dropdown__container__choices__item__unselect-icon"
                    />
                </div>
            </div>
            <p v-if="!bucketsList.length" class="buckets-dropdown__container__no-buckets">
                No Buckets
            </p>
        </div>
    </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import SelectionIcon from '@/../static/images/accessGrants/selection.svg';
import UnselectIcon from '@/../static/images/accessGrants/unselect.svg';

import { ACCESS_GRANTS_ACTIONS } from '@/store/modules/accessGrants';

@Component({
    components: {
        SelectionIcon,
        UnselectIcon,
    },
})
export default class BucketsDropdown extends Vue {
    public bucketSearch = '';

    /**
     * Clears selection of specific buckets and closes dropdown.
     */
    public selectAllBuckets(): void {
        this.$store.dispatch(ACCESS_GRANTS_ACTIONS.CLEAR_SELECTION);
        this.$emit('close');
    }

    /**
     * Toggles bucket selection.
     */
    public toggleBucketSelection(name: string): void {
        this.$store.dispatch(ACCESS_GRANTS_ACTIONS.TOGGLE_BUCKET_SELECTION, name);
    }

    /**
     * Indicates if bucket name is selected.
     * @param name
     */
    public isNameSelected(name: string): boolean {
        return this.selectedBucketNames.includes(name);
    }

    /**
     * Returns stored bucket names list filtered by search string.
     */
    public get bucketsList(): string[] {
        const NON_EXIST_INDEX = -1;
        const buckets: string[] = this.$store.state.bucketUsageModule.allBucketNames;

        return buckets.filter((name: string) => {
            return name.indexOf(this.bucketSearch.toLowerCase()) !== NON_EXIST_INDEX;
        });
    }

    /**
     * Returns stored selected bucket names.
     */
    public get selectedBucketNames(): string[] {
        return this.$store.state.accessGrantsModule.selectedBucketNames;
    }
}
</script>

<style scoped lang="scss">
    ::-webkit-scrollbar,
    ::-webkit-scrollbar-track,
    ::-webkit-scrollbar-thumb {
        margin: 0;
        width: 0;
    }

    .buckets-dropdown {
        position: absolute;
        z-index: 1120;
        left: 0;
        top: calc(100% + 5px);
        box-shadow: 0 20px 34px rgba(10, 27, 44, 0.28);
        border-radius: 6px;
        background-color: #fff;
        border: 1px solid rgba(56, 75, 101, 0.4);
        width: 100%;
        padding: 10px 0;

        &__container {
            overflow-y: scroll;
            overflow-x: hidden;
            height: auto;
            width: 100%;
            max-height: 230px;
            background-color: #fff;
            border-radius: 6px;
            font-family: 'font_regular', sans-serif;
            font-style: normal;
            font-weight: normal;
            font-size: 16px;
            line-height: 21px;
            color: #384b65;

            &__search {
                padding: 5px 10px;
                width: calc(100% - 20px);

                &__input {
                    font-size: 14px;
                    line-height: 18px;
                    border-radius: 6px;
                    width: calc(100% - 30px);
                    padding: 5px;
                }
            }

            &__all {
                margin: 0;
                cursor: pointer;
                background-color: #fff;
                width: calc(100% - 50px);
                padding: 15px 0 15px 50px;

                &:hover {
                    background-color: #ecedf2;
                }
            }

            &__no-buckets {
                font-family: 'font_bold', sans-serif;
                margin: 0;
                font-size: 18px;
                line-height: 24px;
                cursor: default;
                color: #000;
                background-color: #fff;
                width: 100%;
                padding: 15px 0;
                text-align: center;
            }

            &__choices {

                &__item__unselect-icon {
                    opacity: 0;
                }

                .selected {
                    background-color: #f5f6fa;

                    .bucket-name-selection-path {
                        stroke: #0068dc !important;
                    }

                    &:hover {

                        .buckets-dropdown__container__choices__item__unselect-icon {
                            opacity: 1 !important;
                        }
                    }
                }

                &__item {
                    display: flex;
                    align-items: center;
                    justify-content: space-between;
                    padding: 15px 20px;
                    width: calc(100% - 40px);

                    &__left {
                        display: flex;
                        align-items: center;

                        &__label {
                            margin: 0 0 0 15px;
                        }
                    }

                    &:hover {
                        background-color: #ecedf2;

                        .bucket-name-selection-path {
                            stroke: #d4d9e1;
                        }
                    }
                }
            }
        }
    }
</style>
