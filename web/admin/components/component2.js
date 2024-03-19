import store from '../store.js'
import apiCmd from '../apicmd.js'

export default {
    setup(props) {
        const {watchEffect, ref} = Vue;
        const {useConfirm} = primevue.useconfirm;
        const {useToast} = primevue.usetoast;

        const toast = useToast();

        const tabledata = ref(null);
        const uuidV4Create = ref(null)
        const uuidV4Delete = ref(null)

        // Populate synctoken list if valid admintoken
        watchEffect(() => {
            if (store.isAdminTokenValid) {
                getSyncTokens()
            }
        })

        function getSyncTokens() {
            apiCmd.getSyncTokens()
                .then(data => {
                    tabledata.value = data
                })
                .catch(e => {
                    console.error(e)
                })
        }

        const showDialogCreateToken = ref(false);
        function createSyncToken() {
            apiCmd.createSyncToken()
                .then(data => {
                    uuidV4Create.value=data
                    showDialogCreateToken.value=true
                    getSyncTokens()
                })
                .catch(e => {
                    console.error(e)
                    toast.add({severity: 'error', summary: 'createSyncToken error', detail: e, life: 5000});
                })
        }

        const showDialogDeleteToken = ref(false)
        function dialogDeleteToken(token) {
            uuidV4Delete.value=token
            showDialogDeleteToken.value=true
        }

        function deleteSyncToken(token) {
            apiCmd.deleteSyncToken(token)
                .then(data => {
                    getSyncTokens()
                    toast.add({severity: 'success', summary: 'Sync token deleted', detail: token, life: 3000});
                    showDialogDeleteToken.value=false
                })
                .catch(e => {
                    console.error(e)
                    toast.add({severity: 'error', summary: 'deleteSyncToken error', detail: e, life: 5000});
                })

        }

        const confirm = useConfirm();

        function copyToClipboard(text) {
            navigator.clipboard.writeText(text)
            toast.add({severity: 'success', summary: 'Copied to clipboard', detail: text, life: 3000});
        }

        const origin=window.location.origin

        return {
            store,
            tabledata,
            dialogDeleteToken,
            createSyncToken,
            copyToClipboard,
            showDialogCreateToken,
            showDialogDeleteToken,
            uuidV4Create,
            uuidV4Delete,
            deleteSyncToken,
            origin
        };
    },


    template: `<div v-if="store.isAdminTokenValid">

    <div class="flex text-xl my-4">Sync Tokens</div>
    <div class="flex justify-content-end my-2">
        <Button v-on:click="createSyncToken()" outlined icon="pi pi-plus" size="small" label="Create Sync Token"
                severity="success"/>
    </div>
    <div class="flex">

        <DataTable :value="tabledata" class="w-full" size="small" rowHover showGridlines>
            <template #empty>
                <div class="flex my-4 justify-content-center">No Sync Tokens found</div>
            </template>

            <Column field="synctoken" header="Sync Token">
                <template #body="slotProps">
                    <div class="flex flex-nowrap justify-content-left align-content-top gap-2 m-0 p-0">
                        <div>
                            <pre>{{ slotProps.data.synctoken }}</pre>
                        </div>
                        <div class="flex align-items-center">
                            <Button v-on:click="copyToClipboard(slotProps.data.synctoken)" icon="pi pi-copy"
                                    severity="secondary" text rounded size="small"/>
                        </div>
                        <div v-if="slotProps.data.synctoken===uuidV4Create" class="flex align-items-center">
                            <Tag value="New"></Tag>
                        </div>
                    </div>
                </template>
            </Column>
            <Column field="action">
                <template #header><span class="w-full text-center">Action</span></template>
                <template #body="slotProps">
                    <div class="flex flex-nowrap justify-content-center m-0 p-0">
                        <div>
                            <Button v-on:click="dialogDeleteToken(slotProps.data.synctoken)" icon="pi pi-trash"
                                    severity="danger" text size="small" label="Delete"/>
                        </div>
                    </div>
                </template>
            </Column>
        </DataTable>
    </div>

    <!--DELETE TOKEN-->
    <Dialog v-model:visible="showDialogDeleteToken" modal header="Delete Sync Token" :style="{ width: '35rem' }">
        <div class="flex align-items-center gap-3 mb-3">
            <label for="synctoken" class="font-semibold w-6rem">Sync Token</label>
            <InputText v-model="uuidV4Delete" id="synctoken" class="flex-auto" style="text-align: center;" autocomplete="off"
                       disabled/>
        </div>
        <span class="p-text-secondary block mb-5">All Tabby configurations associated with this token will also be deleted.<b>This operation is irreversible.</b></span>
        <div class="flex justify-content-end gap-2 mt-4">
            <Button type="button" label="Cancel" severity="secondary" outlined
                    @click="showDialogDeleteToken = false"></Button>
            <Button type="button" severity="danger" label="Delete" @click="deleteSyncToken(uuidV4Delete)"></Button>
        </div>
    </Dialog>

    <!--ADD TOKEN-->
    <Dialog v-model:visible="showDialogCreateToken" modal header="New Sync Token created" :style="{ width: '45rem' }">
        <span class="p-text-secondary block mb-5">These are the Tabby <b>Config Sync</b> settings for the new sync token </span>
        <div class="flex align-items-center gap-3 mb-3">
            <label for="synchost" class="font-semibold w-9rem">Sync Host</label>
            <InputText :placeholder="origin" id="synchost" class="flex-auto" style="text-align: center; width: 20em"
                       autocomplete="off" disabled/>
            <div>
                <Button v-on:click="copyToClipboard(origin)" icon="pi pi-copy" severity="secondary" text rounded
                        size="large"/>
            </div>
        </div>
        <div class="flex align-items-center gap-3 mb-3">
            <label for="synctoken" class="font-semibold w-9rem">Secret&nbsp;sync&nbsp;Token</label>
            <InputText v-model="uuidV4Create" id="synctoken" class="flex-auto" style="text-align: center;" autocomplete="off"
                       disabled/>
            <div>
                <Button v-on:click="copyToClipboard(uuidV4Create)" icon="pi pi-copy" severity="secondary" text rounded
                        size="large"/>
            </div>
        </div>
        <div class="flex justify-content-end gap-2 mt-4">
            <Button type="button" label="OK" @click="showDialogCreateToken = false"></Button>
        </div>
    </Dialog>
</div>
    `,
};