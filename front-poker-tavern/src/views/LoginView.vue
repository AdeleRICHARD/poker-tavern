<template>
    <div class="login-view">
        <div class="ambient-glow" aria-hidden="true"></div>

        <div class="login-container">
            <header class="page-header">
                <div class="sign-assembly">
                    <div class="iron-hanger" aria-hidden="true">
                        <span class="iron-bar"></span>
                        <span class="iron-scroll"></span>
                    </div>
                    <span class="sign-chain sign-chain--left" aria-hidden="true"></span>
                    <span class="sign-chain sign-chain--right" aria-hidden="true"></span>
                    <div class="tavern-sign">
                        <span class="beer-mark" aria-hidden="true">🍺</span>
                        <span class="sign-copy">
                            <small>Bienvenue à</small>
                            <strong>La Taverne du Planning Poker</strong>
                        </span>
                    </div>
                </div>
            </header>

            <section class="tabletop-layout" aria-label="Accès et déroulement d’une session">
                <section class="login-card" aria-labelledby="join-title">
                    <div class="card-heading">
                        <div>
                            <p class="card-kicker">Votre table vous attend</p>
                            <h2 id="join-title">Rejoindre une session</h2>
                        </div>
                        <span class="card-icon" aria-hidden="true">✦</span>
                    </div>

                    <form @submit.prevent="handleLogin" class="login-form">
                        <div class="form-group">
                            <label for="playerName">Votre nom</label>
                            <div class="input-shell">
                                <span aria-hidden="true">♙</span>
                                <input
                                    id="playerName"
                                    v-model="playerName"
                                    type="text"
                                    placeholder="Saisissez votre nom"
                                    autocomplete="name"
                                    required
                                />
                            </div>
                        </div>

                        <div class="form-group">
                            <div class="label-row">
                                <label for="sessionId">Identifiant de session</label>
                                <span>Facultatif</span>
                            </div>
                            <div class="input-shell">
                                <span aria-hidden="true">⌘</span>
                                <input
                                    id="sessionId"
                                    v-model="sessionId"
                                    type="text"
                                    placeholder="Ex. : A7K9M2"
                                    autocomplete="off"
                                />
                            </div>
                            <div v-if="sessionId" class="session-prefilled-notice">
                                <span aria-hidden="true">✓</span>
                                Session trouvée : il ne manque plus que votre nom.
                            </div>
                        </div>

                        <button
                            type="submit"
                            class="primary-button"
                            :disabled="!playerName.trim() || isLoading"
                        >
                            <span>{{ isLoading ? "Connexion..." : "Rejoindre la table" }}</span>
                            <span aria-hidden="true">→</span>
                        </button>
                    </form>

                    <div class="login-divider">
                        <span>ou créez votre propre table</span>
                    </div>

                    <form @submit.prevent="handleCreateSession" class="create-form">
                        <div class="form-group create-field">
                            <label for="newSessionName">Nom de la nouvelle session</label>
                            <div class="input-shell">
                                <span aria-hidden="true">◇</span>
                                <input
                                    id="newSessionName"
                                    v-model="newSessionName"
                                    type="text"
                                    placeholder="Ex. : Planification du sprint 24"
                                    autocomplete="off"
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            class="secondary-button"
                            :disabled="
                                !playerName.trim() ||
                                !newSessionName.trim() ||
                                isLoading
                            "
                        >
                            {{ isLoading ? "Création..." : "Créer la session" }}
                        </button>
                    </form>

                    <div v-if="errorMessage" class="error-message" role="alert">
                        <span aria-hidden="true">!</span>
                        {{ errorMessage }}
                    </div>
                </section>

                <aside class="ritual-card" aria-label="Déroulement d’une estimation">
                    <div class="ritual-heading">
                        <p>En trois étapes</p>
                        <h2>Comment ça marche ?</h2>
                    </div>
                    <ol class="ritual-steps">
                        <li>
                            <span>01</span>
                            <div>
                                <strong>Rassemblez</strong>
                                <p>Invitez l’équipe avec un simple lien.</p>
                            </div>
                        </li>
                        <li>
                            <span>02</span>
                            <div>
                                <strong>Estimez</strong>
                                <p>Votez simultanément, sans vous influencer.</p>
                            </div>
                        </li>
                        <li>
                            <span>03</span>
                            <div>
                                <strong>Alignez-vous</strong>
                                <p>Révélez les cartes et décidez ensemble.</p>
                            </div>
                        </li>
                    </ol>
                </aside>
            </section>

            <blockquote class="table-quote">
                <span aria-hidden="true">“</span>
                <p>Les meilleures estimations commencent autour d’une bonne table.</p>
            </blockquote>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useGameStore } from "@/stores/gameStore";

// Emit event to parent component when login is successful
const emit = defineEmits<{
    loginSuccess: [];
}>();

// Store
const gameStore = useGameStore();

// Form data
const playerName = ref("");
const sessionId = ref("");
const newSessionName = ref("");
const isLoading = ref(false);
const errorMessage = ref("");

// Check URL parameters for invite link
onMounted(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const sessionIdParam = urlParams.get("sessionId");
    if (sessionIdParam) {
        sessionId.value = sessionIdParam;
        // Clean up URL after extracting the parameter
        window.history.replaceState(
            {},
            document.title,
            window.location.pathname,
        );
    }
});

// Handlers
async function handleLogin() {
    if (!playerName.value.trim()) return;

    isLoading.value = true;
    errorMessage.value = "";

    try {
        if (sessionId.value.trim()) {
            // Join existing session
            await gameStore.joinSession(
                sessionId.value.trim(),
                playerName.value.trim(),
            );
        } else {
            // Create default session
            await gameStore.createSession(
                `Session de ${playerName.value}`,
                [],
                playerName.value.trim(),
            );
        }

        emit("loginSuccess");
    } catch (error) {
        console.error("Login error:", error);
        if (error instanceof Error) {
            if (error.message.includes("Session not found")) {
                errorMessage.value =
                    "Session introuvable. Vérifiez son identifiant ou créez une nouvelle session.";
            } else if (error.message.includes("404")) {
                errorMessage.value =
                    "Session introuvable. Vérifiez son identifiant.";
            } else {
                errorMessage.value = "Impossible de se connecter. Veuillez réessayer.";
            }
        } else {
            errorMessage.value = "Impossible de se connecter. Veuillez réessayer.";
        }
    } finally {
        isLoading.value = false;
    }
}

async function handleCreateSession() {
    if (!playerName.value.trim() || !newSessionName.value.trim()) return;

    isLoading.value = true;
    errorMessage.value = "";

    try {
        // Create new session
        await gameStore.createSession(
            newSessionName.value.trim(),
            [],
            playerName.value.trim(),
        );

        emit("loginSuccess");
    } catch (error) {
        errorMessage.value = "Impossible de créer la session. Veuillez réessayer.";
        console.error("Create session error:", error);
    } finally {
        isLoading.value = false;
    }
}
</script>

<style scoped>
.login-view {
    position: relative;
    isolation: isolate;
    min-height: 100dvh;
    width: 100%;
    overflow: auto;
    padding: clamp(1.5rem, 4vw, 3.5rem);
    background-image:
        linear-gradient(90deg, rgba(16, 9, 5, 0.94) 0%, rgba(16, 9, 5, 0.76) 38%, rgba(16, 9, 5, 0.18) 72%, rgba(16, 9, 5, 0.42) 100%),
        linear-gradient(0deg, rgba(11, 6, 3, 0.82) 0%, transparent 35%),
        url("../assets/tavern-home-bg.webp");
    background-position: center;
    background-size: cover;
    color: #fff8ea;
    font-family: "Lato", sans-serif;
}

.login-view::before {
    position: absolute;
    inset: 0;
    z-index: -1;
    background: radial-gradient(circle at 72% 42%, rgba(219, 122, 43, 0.16), transparent 30%);
    content: "";
    pointer-events: none;
}

.ambient-glow {
    position: absolute;
    top: 11%;
    left: 8%;
    z-index: -1;
    width: min(32rem, 70vw);
    height: min(32rem, 70vw);
    border-radius: 50%;
    background: rgba(197, 103, 39, 0.12);
    filter: blur(6rem);
    pointer-events: none;
}

.login-container {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    min-height: calc(100dvh - 7rem);
    width: min(100%, 78rem);
    margin: 0 auto;
}

.page-header {
    display: flex;
    justify-content: center;
    margin-top: calc(0rem - clamp(1.5rem, 4vw, 3.5rem));
}

.sign-assembly {
    position: relative;
    display: grid;
    place-items: center;
    width: min(100%, 46rem);
    padding-top: 5.4rem;
}

.iron-hanger {
    position: absolute;
    top: 0.35rem;
    left: 50%;
    width: calc(100% - 2rem);
    height: 3.2rem;
    transform: translateX(-50%);
}

.iron-bar {
    position: absolute;
    top: 1.5rem;
    left: 0;
    width: 100%;
    height: 0.72rem;
    border: 1px solid #120d0a;
    border-radius: 999rem;
    background: linear-gradient(180deg, #7a6959 0%, #302720 38%, #100d0b 72%, #514238 100%);
    box-shadow:
        0 0.22rem 0.3rem rgba(0, 0, 0, 0.75),
        inset 0 1px rgba(255, 221, 175, 0.2);
}

.iron-bar::before,
.iron-bar::after {
    position: absolute;
    top: 50%;
    width: 1.6rem;
    height: 1.15rem;
    border: 1px solid #17100c;
    border-radius: 55% 16% 16% 55%;
    background: linear-gradient(180deg, #786552, #211a15 58%, #57473b);
    box-shadow: inset 0 1px rgba(255, 222, 178, 0.18);
    content: "";
    transform: translateY(-50%);
}

.iron-bar::before {
    left: -0.65rem;
}

.iron-bar::after {
    right: -0.65rem;
    transform: translateY(-50%) rotate(180deg);
}

.iron-scroll {
    position: absolute;
    top: 0.05rem;
    left: 50%;
    width: 8.5rem;
    height: 2.15rem;
    border-top: 0.35rem solid #302720;
    border-radius: 50% 50% 0 0;
    filter: drop-shadow(0 0.15rem 0.15rem rgba(0, 0, 0, 0.85));
    transform: translateX(-50%);
}

.iron-scroll::before,
.iron-scroll::after {
    position: absolute;
    top: -0.25rem;
    width: 1.65rem;
    height: 1.65rem;
    border: 0.3rem solid #40352c;
    border-radius: 50%;
    content: "";
}

.iron-scroll::before {
    left: 1.25rem;
}

.iron-scroll::after {
    right: 1.25rem;
}

.tavern-sign {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-self: center;
    gap: 0.9rem;
    align-items: center;
    width: fit-content;
    max-width: 100%;
    margin: 0;
    padding: 0.8rem 1.25rem 0.8rem 0.9rem;
    border: 2px solid #8f5a30;
    border-radius: 0.45rem;
    background:
        radial-gradient(ellipse at 22% 34%, transparent 0 0.42rem, rgba(40, 19, 9, 0.46) 0.48rem 0.56rem, transparent 0.62rem),
        radial-gradient(ellipse at 78% 70%, transparent 0 0.3rem, rgba(35, 16, 8, 0.4) 0.36rem 0.43rem, transparent 0.5rem),
        repeating-linear-gradient(176deg, rgba(255, 218, 160, 0.045) 0 0.08rem, transparent 0.08rem 0.48rem),
        linear-gradient(90deg, rgba(255, 210, 144, 0.13), transparent 32%, rgba(0, 0, 0, 0.22)),
        linear-gradient(180deg, #704326, #472514 52%, #30180e);
    color: #fff0d5;
    box-shadow:
        0 0.8rem 1.8rem rgba(0, 0, 0, 0.38),
        inset 0 0 0 1px rgba(255, 220, 166, 0.14),
        inset 0 -0.7rem 1.2rem rgba(16, 7, 3, 0.28);
    transform: rotate(-0.6deg);
}

.tavern-sign::after {
    position: absolute;
    inset: 0.3rem;
    border: 1px solid rgba(231, 172, 101, 0.24);
    border-radius: 0.25rem;
    content: "";
    pointer-events: none;
}

.sign-chain {
    position: absolute;
    top: 1.85rem;
    width: 0.68rem;
    height: 3.7rem;
    background: radial-gradient(
            ellipse at center,
            transparent 28%,
            #4b2f1d 31% 39%,
            #c0925d 42% 56%,
            #4b2f1d 59% 65%,
            transparent 68%
        )
        center top / 0.68rem 0.78rem repeat-y;
    filter: drop-shadow(0 0.12rem 0.12rem rgba(0, 0, 0, 0.9));
    opacity: 0.95;
    transform-origin: bottom;
}

.sign-chain--left {
    left: 28%;
    transform: rotate(8deg);
}

.sign-chain--right {
    right: 28%;
    transform: rotate(-8deg);
}

.beer-mark {
    display: grid;
    place-items: center;
    width: 2.55rem;
    height: 2.55rem;
    border: 1px solid rgba(240, 185, 112, 0.35);
    border-radius: 50%;
    background: rgba(29, 13, 7, 0.34);
    font-size: 1.45rem;
    filter: sepia(0.18) saturate(0.92);
}

.sign-copy {
    display: grid;
    gap: 0.08rem;
}

.sign-copy small {
    color: #dca869;
    font-size: 0.58rem;
    font-weight: 700;
    letter-spacing: 0.16em;
    text-transform: uppercase;
}

.sign-copy strong {
    font-family: "Cinzel", serif;
    font-size: 1.08rem;
    font-weight: 600;
    line-height: 1.2;
    text-shadow: 0 0.125rem 0.35rem rgba(0, 0, 0, 0.6);
}

.card-kicker,
.ritual-heading > p {
    color: #e9b36f;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
}

.login-card,
.ritual-card {
    border: 1px solid rgba(255, 236, 207, 0.16);
    background: linear-gradient(145deg, rgba(40, 23, 14, 0.92), rgba(20, 12, 8, 0.88));
    box-shadow: 0 1.5rem 4rem rgba(0, 0, 0, 0.42), inset 0 1px rgba(255, 255, 255, 0.04);
    backdrop-filter: blur(1.25rem) saturate(115%);
}

.login-card {
    padding: clamp(1.25rem, 3vw, 1.8rem);
    border-radius: 1.35rem;
}

.tabletop-layout {
    display: grid;
    grid-template-columns: minmax(31rem, 35rem) minmax(20rem, 25rem);
    gap: clamp(2rem, 3vw, 3rem);
    justify-content: center;
    align-items: stretch;
    width: min(100%, 63rem);
    margin-inline: auto;
    margin-top: auto;
}

.card-heading {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.35rem;
}

.card-heading h2,
.ritual-heading h2 {
    margin-top: 0.25rem;
    color: #fff8ea;
    font-family: "Cinzel", serif;
    font-size: 1.3rem;
    font-weight: 600;
}

.card-icon {
    display: grid;
    place-items: center;
    width: 2.4rem;
    height: 2.4rem;
    border: 1px solid rgba(232, 177, 103, 0.32);
    border-radius: 0.8rem;
    background: rgba(232, 177, 103, 0.08);
    color: #e8b167;
}

.login-form {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.85rem;
}

.form-group {
    min-width: 0;
}

.form-group label {
    display: block;
    margin-bottom: 0.45rem;
    color: rgba(255, 248, 234, 0.82);
    font-size: 0.73rem;
    font-weight: 700;
    letter-spacing: 0.02em;
}

.label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.label-row > span {
    color: rgba(255, 248, 234, 0.42);
    font-size: 0.65rem;
}

.input-shell {
    display: flex;
    gap: 0.65rem;
    align-items: center;
    min-height: 3rem;
    padding: 0 0.85rem;
    border: 1px solid rgba(255, 236, 207, 0.13);
    border-radius: 0.8rem;
    background: rgba(10, 7, 5, 0.5);
    color: rgba(232, 177, 103, 0.64);
    transition: border-color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.input-shell:focus-within {
    border-color: rgba(232, 177, 103, 0.62);
    background: rgba(12, 8, 5, 0.72);
    box-shadow: 0 0 0 0.2rem rgba(232, 177, 103, 0.08);
}

.input-shell input {
    min-width: 0;
    width: 100%;
    border: 0;
    outline: 0;
    background: transparent;
    color: #fff8ea;
    font: inherit;
    font-size: 0.85rem;
}

.input-shell input::placeholder {
    color: rgba(255, 248, 234, 0.32);
}

.primary-button,
.secondary-button {
    display: flex;
    justify-content: center;
    gap: 0.7rem;
    align-items: center;
    min-height: 3rem;
    border-radius: 0.8rem;
    font-family: "Lato", sans-serif;
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    cursor: pointer;
    transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.primary-button {
    grid-column: 1 / -1;
    border: 1px solid #e2a45e;
    background: linear-gradient(135deg, #d78c43, #b9662d);
    color: #1d0f08;
    box-shadow: 0 0.6rem 1.5rem rgba(181, 91, 34, 0.22);
}

.primary-button:hover:not(:disabled),
.secondary-button:hover:not(:disabled) {
    transform: translateY(-0.1rem);
    box-shadow: 0 0.8rem 1.8rem rgba(181, 91, 34, 0.3);
}

.primary-button:focus-visible,
.secondary-button:focus-visible {
    outline: 2px solid #fff2dc;
    outline-offset: 0.2rem;
}

.primary-button:disabled,
.secondary-button:disabled {
    opacity: 0.38;
    cursor: not-allowed;
}

.login-divider {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    margin: 1.15rem 0 0.9rem;
    color: rgba(255, 248, 234, 0.4);
    font-size: 0.65rem;
    letter-spacing: 0.04em;
    text-transform: uppercase;
}

.login-divider::before,
.login-divider::after {
    flex: 1;
    height: 1px;
    background: rgba(255, 236, 207, 0.12);
    content: "";
}

.create-form {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.8rem;
    align-items: end;
}

.secondary-button {
    min-width: 9rem;
    border: 1px solid rgba(232, 177, 103, 0.35);
    background: rgba(232, 177, 103, 0.07);
    color: #f0bd78;
}

.secondary-button:hover:not(:disabled) {
    background: rgba(232, 177, 103, 0.13);
}

.session-prefilled-notice,
.error-message {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-top: 0.55rem;
    padding: 0.55rem 0.7rem;
    border-radius: 0.65rem;
    font-size: 0.7rem;
}

.session-prefilled-notice {
    border: 1px solid rgba(114, 187, 134, 0.22);
    background: rgba(64, 127, 82, 0.12);
    color: #a8d6b2;
}

.error-message {
    border: 1px solid rgba(232, 103, 82, 0.32);
    background: rgba(130, 46, 34, 0.2);
    color: #ffc2b7;
}

.error-message > span {
    display: grid;
    place-items: center;
    width: 1.2rem;
    height: 1.2rem;
    border-radius: 50%;
    background: rgba(232, 103, 82, 0.18);
    font-weight: 700;
}

.table-quote {
    position: relative;
    width: min(100%, 63rem);
    max-width: none;
    margin: clamp(1.5rem, 3vh, 2.5rem) auto auto;
    padding: 1.5rem 3rem;
    border-left: 2px solid rgba(228, 166, 94, 0.62);
    background: linear-gradient(90deg, rgba(24, 13, 8, 0.68), rgba(24, 13, 8, 0.3));
    text-shadow: 0 0.125rem 1rem rgba(0, 0, 0, 0.72);
    backdrop-filter: blur(0.25rem);
}

.table-quote > span {
    position: absolute;
    top: 0.3rem;
    left: 0.8rem;
    color: rgba(228, 166, 94, 0.65);
    font-family: Georgia, serif;
    font-size: 3rem;
    line-height: 1;
}

.table-quote p {
    max-width: 46rem;
    margin: 0 auto;
    color: rgba(255, 248, 234, 0.9);
    font-family: "Cinzel", serif;
    font-size: clamp(1.25rem, 2.2vw, 1.8rem);
    font-weight: 500;
    line-height: 1.4;
    text-align: center;
}

.ritual-card {
    align-self: stretch;
    width: 100%;
    padding: 1.5rem;
    border-radius: 1.2rem;
}

.ritual-steps {
    display: grid;
    gap: 0.35rem;
    margin-top: 1rem;
    list-style: none;
}

.ritual-steps li {
    display: grid;
    grid-template-columns: 2.2rem 1fr;
    gap: 1rem;
    padding: 1rem 0;
    border-top: 1px solid rgba(255, 236, 207, 0.1);
}

.ritual-steps li > span {
    padding-top: 0.15rem;
    color: #e2a45e;
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.08em;
}

.ritual-steps strong {
    color: #fff8ea;
    font-family: "Cinzel", serif;
    font-size: 1rem;
}

.ritual-steps p {
    margin-top: 0.2rem;
    color: rgba(255, 248, 234, 0.55);
    font-size: 0.82rem;
    line-height: 1.55;
}

@media (max-width: 68rem) {
    .tabletop-layout {
        grid-template-columns: minmax(28rem, 34rem) minmax(19rem, 23rem);
        gap: 2rem;
        width: min(100%, 59rem);
    }

    .table-quote {
        width: min(100%, 59rem);
    }
}

@media (max-width: 52rem) {
    .login-view {
        padding: 1.5rem;
        background-position: 62% center;
    }

    .login-view::after {
        position: absolute;
        inset: 0;
        z-index: -1;
        background: rgba(14, 8, 5, 0.42);
        content: "";
    }

    .login-container {
        min-height: auto;
    }

    .page-header {
        margin-top: -1.5rem;
    }

    .sign-assembly {
        padding-top: 5.2rem;
    }

    .tabletop-layout {
        grid-template-columns: 1fr;
        gap: 1.5rem;
        width: min(100%, 38rem);
        margin: 0.5rem auto 0;
    }

    .table-quote {
        width: min(100%, 38rem);
        margin: 1.5rem auto 0;
        padding: 1.25rem 2.5rem;
    }

    .ritual-card {
        width: min(100%, 38rem);
    }
}

@media (max-width: 38rem) {
    .login-view {
        padding: 1rem;
        background-position: 68% center;
    }

    .tavern-sign {
        padding-right: 0.9rem;
    }

    .page-header {
        margin-top: -1rem;
    }

    .iron-hanger {
        width: calc(100% - 1.5rem);
    }

    .iron-scroll {
        transform: translateX(-50%) scale(0.82);
    }

    .sign-copy strong {
        font-size: 0.9rem;
    }

    .login-card {
        padding: 1.15rem;
        border-radius: 1.1rem;
    }

    .login-form,
    .create-form {
        grid-template-columns: 1fr;
    }

    .secondary-button {
        width: 100%;
    }

    .ritual-card {
        padding: 1.15rem;
    }
}
</style>
