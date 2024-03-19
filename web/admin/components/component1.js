import store from '../store.js'
import apiCmd from '../apicmd.js'

export default {
    setup() {
        const {watchEffect} = Vue;

        watchEffect(() => {
            apiCmd.getSyncTokens()
                .then(data => {
                    store.isAdminTokenValid = true
                })
                .catch(e => {
                    console.error(e)
                    store.isAdminTokenValid = false
                })
        })

        return {store};
    },

    template: `<div>
    <div class="flex text-xl my-4">Admin Token</div>
    <div class="flex gap-2">
        <InputText style="width: 25em;" v-model="store.$adminToken" id="admintoken" placeholder="Enter Admin Token"
                   :invalid="!store.isAdminTokenValid" type="password" variant="filled"/>
        <InlineMessage :severity="store.isAdminTokenValid?'success':'error'">
            {{ store.isAdminTokenValid ? "Valid Admin Token" : "Valid Admin Token is required" }}
        </InlineMessage>
    </div>
    </div>
    `,
};