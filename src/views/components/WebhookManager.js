export default {
    name: 'WebhookManager',
    data() {
        return {
            loading: false,
            webhooks: [],
            showAddForm: false,
            editingWebhookId: null,
            allEvents: [],
            webhookForm: {
                url: '',
                secret: '',
                events: [],
                enabled: true,
                description: ''
            }
        }
    },
    methods: {
        async fetchAvailableEvents() {
            try {
                const response = await window.http.get('/webhook/events')
                this.allEvents = response.data.results
            } catch (error) {
                console.error('Failed to fetch available events:', error)
                showErrorInfo('Failed to load available events')
            }
        },
        openModal() {
            this.fetchWebhooks()
            this.fetchAvailableEvents()
            this.showAddForm = false
            $('#modalWebhookManager').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        async fetchWebhooks() {
            try {
                const response = await window.http.get('/webhook')
                this.webhooks = response.data.results
            } catch (error) {
                console.error('Failed to fetch webhooks:', error)
                showErrorInfo('Failed to load webhooks')
            }
        },
        isValidForm() {
            if (!this.webhookForm.url || !this.webhookForm.url.trim()) {
                return false
            }
            
            if (this.webhookForm.events.length === 0) {
                return false
            }
            
            return true
        },
        toggleEvent(event) {
            const index = this.webhookForm.events.indexOf(event)
            if (index > -1) {
                this.webhookForm.events.splice(index, 1)
            } else {
                this.webhookForm.events.push(event)
            }
        },
        isEventSelected(event) {
            return this.webhookForm.events.includes(event)
        },
        getSelectedEventsText() {
            if (this.webhookForm.events.length === 0) {
                return 'None'
            }
            return this.webhookForm.events.join(', ')
        },
        async handleSubmit() {
            if (!this.isValidForm() || this.loading) {
                return;
            }
            try {
                if (this.editingWebhookId) {
                    await this.updateApi()
                    showSuccessInfo('Webhook updated successfully')
                } else {
                    await this.submitApi()
                    showSuccessInfo('Webhook created successfully')
                }
                this.handleReset()
                this.showAddForm = false
                this.editingWebhookId = null
                await this.fetchWebhooks()
            } catch (err) {
                showErrorInfo(err)
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                await window.http.post('/webhook', this.webhookForm)
            } catch (error) {
                if (error.response) {
                    throw new Error(error.response.data.message);
                }
                throw new Error(error.message);
            } finally {
                this.loading = false;
            }
        },
        handleReset() {
            this.webhookForm = {
                url: '',
                secret: '',
                events: [],
                enabled: true,
                description: ''
            }
            this.editingWebhookId = null
        },

        async updateApi() {
            this.loading = true;
            try {
                await window.http.put(`/webhook/${this.editingWebhookId}`, this.webhookForm)
            } catch (error) {
                if (error.response) {
                    throw new Error(error.response.data.message);
                }
                throw new Error(error.message);
            } finally {
                this.loading = false;
            }
        },

        editWebhook(webhook) {
            this.webhookForm = {
                url: webhook.url,
                secret: webhook.secret || '',
                events: [...webhook.events],
                enabled: webhook.enabled,
                description: webhook.description || ''
            }
            this.editingWebhookId = webhook.id
            this.showAddForm = true
            
            this.$nextTick(() => {
                if (this.$refs.urlInput) {
                    this.$refs.urlInput.focus()
                }
            })
        },

        toggleAddForm() {
            this.showAddForm = !this.showAddForm
            if (!this.showAddForm) {
                this.handleReset()
            } else {
                this.$nextTick(() => {
                    if (this.$refs.urlInput) {
                        this.$refs.urlInput.focus()
                    }
                })
            }
        },
        async deleteWebhook(id) {
            if (!confirm('Are you sure you want to delete this webhook?')) {
                return
            }
            
            try {
                await window.http.delete(`/webhook/${id}`)
                showSuccessInfo('Webhook deleted successfully')
                await this.fetchWebhooks()
            } catch (error) {
                showErrorInfo('Failed to delete webhook')
            }
        },
        
        async toggleWebhookEnabled(webhook) {
            try {
                const updatedWebhook = {
                    ...webhook,
                    enabled: !webhook.enabled
                };
                
                await window.http.put(`/webhook/${webhook.id}`, updatedWebhook)
                showSuccessInfo(`Webhook ${updatedWebhook.enabled ? 'enabled' : 'disabled'} successfully`)
                await this.fetchWebhooks()
            } catch (error) {
                showErrorInfo('Failed to update webhook status')
                await this.fetchWebhooks()
            }
        }
    },
    template: `
    <div class="green card" @click="openModal" style="cursor: pointer">
        <div class="content">
            <a class="ui teal right ribbon label">App</a>
            <div class="header">Manage Webhooks</div>
            <div class="description">
                Configure webhook endpoints for event notifications
            </div>
        </div>
    </div>
    
    <!-- Webhook Manager Modal -->
    <div class="ui modal" id="modalWebhookManager">
        <i class="close icon"></i>
        <div class="header">
            Manage Webhooks
        </div>
        <div class="scrolling content">
            <transition name="fade" mode="out-in">
                <div v-if="!showAddForm" key="webhook-list">
                    <div class="ui segments" v-if="webhooks.length > 0">
                        <div class="ui segment" v-for="webhook in webhooks" :key="webhook.id">
                            <div class="ui grid middle aligned">
                                <div class="twelve wide column">
                                    <div class="ui small header" style="margin-bottom: 0.5rem;">
                                        <i class="globe icon"></i>
                                        {{ webhook.url }}
                                    </div>
                                    <div class="ui list horizontal" style="margin: 0.5rem 0;">
                                        <div class="toggle-switch-container" style="margin: 0.5rem 0;">
                                            <label class="toggle-switch">
                                                <input type="checkbox" :checked="webhook.enabled" @change="toggleWebhookEnabled(webhook)">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">{{ webhook.enabled ? 'Active' : 'Inactive' }}</span>
                                        </div>
                                    </div>
                                    <div>
                                        <div class="event-cards">
                                            <div v-for="event in allEvents" :key="event"
                                                 class="event-card"
                                                 :class="{'active-event': webhook.events.includes(event), 'inactive-event': !webhook.events.includes(event)}">
                                                {{ event }}
                                            </div>
                                        </div>
                                    </div>
                                    <p v-if="webhook.description" class="ui description text" style="margin-top: 0.5rem; color: #666;">
                                        <i class="file alternate icon"></i>{{ webhook.description }}
                                    </p>
                                </div>
                                <div class="four wide column right aligned">
                                    <div class="ui buttons vertical">
                                        <button class="ui blue compact icon button" style="margin-bottom: 0.5rem;" @click="editWebhook(webhook)" title="Edit webhook">
                                            <i class="edit icon"></i>
                                        </button>
                                        <button class="ui red compact icon button" @click="deleteWebhook(webhook.id)" title="Delete webhook">
                                            <i class="trash alternate icon"></i>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="ui message info" v-else>
                        <div class="header">
                            <i class="info circle icon"></i>
                            No webhooks configured
                        </div>
                        <p>You haven't configured any webhooks yet.</p>
                    </div>
                    
                    <div class="ui divider"></div>
                    
                    <div class="ui form">
                        <div class="field">
                            <button class="ui primary fluid button" @click="toggleAddForm">
                                <i class="plus icon"></i>
                                Add New Webhook
                            </button>
                        </div>
                    </div>
                </div>
                
                <div v-else key="add-form">
                    <div class="ui form">
                        <h5 class="ui dividing header">
                            {{ editingWebhookId ? 'Edit Webhook' : 'Create New Webhook' }}
                        </h5>
                        
                        <div class="field">
                            <label>Webhook URL</label>
                            <input v-model="webhookForm.url" type="url"
                                   placeholder="https://your-webhook-endpoint.com/callback"
                                   aria-label="Webhook URL" ref="urlInput">
                        </div>
                        
                        <div class="field">
                            <label>Secret Key (optional)</label>
                            <input v-model="webhookForm.secret" type="text"
                                   placeholder="Secret for HMAC signature"
                                   aria-label="Secret Key">
                            <div class="ui pointing label">
                                Leave empty to use default secret
                            </div>
                        </div>
                        
                        <div class="field">
                            <label>Events to Receive</label>
                            <div class="ui six column grid">
                                <div class="column" v-for="event in allEvents">
                                    <div class="ui checkbox">
                                        <input type="checkbox" :id="'event-' + event"
                                               :value="event" v-model="webhookForm.events">
                                        <label :for="'event-' + event">{{ event }}</label>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        
                        <div class="field">
                            <label>Description (optional)</label>
                            <textarea v-model="webhookForm.description"
                                      placeholder="Description for this webhook configuration"
                                      rows="2" aria-label="Description"></textarea>
                        </div>
                        
                        <div class="field">
                            <button class="ui primary button" :class="{'loading': loading}"
                                    @click="handleSubmit" type="button" :disabled="loading">
                                {{ editingWebhookId ? 'Update Webhook' : 'Create Webhook' }}
                            </button>
                            <button class="ui button" @click="toggleAddForm" :disabled="loading">
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            </transition>
        </div>
        <div class="actions">
            <div class="ui black deny button">
                Close
            </div>
        </div>
    </div>
    `
}