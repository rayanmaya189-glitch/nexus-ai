# 14. Navigation

## Table of Contents

1. [Navigation Compose Architecture](#1-navigation-compose-architecture)
2. [Navigation Graph](#2-navigation-graph)
3. [Route Definitions](#3-route-definitions)
4. [Route Arguments](#4-route-arguments)
5. [Screen Transitions](#5-screen-transitions)
6. [Bottom Navigation](#6-bottom-navigation)
7. [Navigation Bar Items](#7-navigation-bar-items)
8. [Top App Bar](#8-top-app-bar)
9. [Navigation Drawer](#9-navigation-drawer)
10. [Navigation Actions](#10-navigation-actions)
11. [Deep Links](#11-deep-links)
12. [Deep Link Handling](#12-deep-link-handling)
13. [Back Stack Management](#13-back-stack-management)
14. [Navigation State Preservation](#14-navigation-state-preservation)
15. [Nested Navigation](#15-nested-navigation)
16. [Modal Navigation](#16-modal-navigation)
17. [Sheet Navigation](#17-sheet-navigation)
18. [Navigation Animations](#18-navigation-animations)
19. [Navigation Testing](#19-navigation-testing)
20. [Navigation Accessibility](#20-navigation-accessibility)
21. [Navigation Performance](#21-navigation-performance)
22. [Navigation Error Handling](#22-navigation-error-handling)
23. [Navigation from Notifications](#23-navigation-from-notifications)
24. [Navigation from Share Intent](#24-navigation-from-share-intent)
25. [Navigation from Widget](#25-navigation-from-widget)

---

## 1. Navigation Compose Architecture

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Navigation Architecture                   │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   Activity / NavHost                  │  │
│  │                                                       │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │                    NavHost                       │  │  │
│  │  │                                                  │  │  │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │  │  │
│  │  │  │   Home   │  │   Chat   │  │ Document │     │  │  │
│  │  │  │  Screen  │  │  Screen  │  │  Screen  │     │  │  │
│  │  │  └──────────┘  └──────────┘  └──────────┘     │  │  │
│  │  │                                                  │  │  │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │  │  │
│  │  │  │  Agent   │  │ Settings │  │  Login   │     │  │  │
│  │  │  │  Screen  │  │  Screen  │  │  Screen  │     │  │  │
│  │  │  └──────────┘  └──────────┘  └──────────┘     │  │  │
│  │  │                                                  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                       │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │              Bottom Navigation                   │  │  │
│  │  │  [Dashboard] [Chat] [Agents] [Knowledge] [More] │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  Navigation Controller                 │  │
│  │  • navigate(route)                                    │  │
│  │  • popBackStack()                                     │  │
│  │  • navigateUp()                                       │  │
│  │  • currentBackStackEntryAsState()                     │  │
│  │  • addOnDestinationChangedListener()                  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Core Navigation Setup

```kotlin
// ─── MainActivity ──────────────────────────────────────────

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            NexusTheme {
                val navController = rememberNavController()
                val navBackStackEntry by navController.currentBackStackEntryAsState()
                val currentRoute = navBackStackEntry?.destination?.route

                Scaffold(
                    bottomBar = {
                        if (currentRoute in bottomBarRoutes) {
                            NexusBottomBar(navController = navController)
                        }
                    },
                    topBar = {
                        if (currentRoute in topBarRoutes) {
                            NexusTopBar(navController = navController)
                        }
                    }
                ) { innerPadding ->
                    NexusNavHost(
                        navController = navController,
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }
}

// ─── NavHost ───────────────────────────────────────────────

@Composable
fun NexusNavHost(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    NavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD,
        modifier = modifier,
        enterTransition = { fadeIn() + slideInHorizontally() },
        exitTransition = { fadeOut() + slideOutHorizontally() },
        popEnterTransition = { fadeIn() + slideInHorizontally { -it / 3 } },
        popExitTransition = { fadeOut() + slideOutHorizontally { it / 3 } }
    ) {
        // Auth flows
        authGraph(navController)

        // Main app flows
        dashboardGraph(navController)
        chatGraph(navController)
        agentsGraph(navController)
        knowledgeGraph(navController)
        settingsGraph(navController)

        // Modal flows
        profileGraph(navController)
        searchGraph(navController)
    }
}
```

### Navigation Dependency

```kotlin
// build.gradle.kts
dependencies {
    implementation("androidx.navigation:navigation-compose:2.7.7")
    implementation("androidx.hilt:hilt-navigation-compose:1.2.0")
    implementation("androidx.lifecycle:lifecycle-runtime-compose:2.7.0")
}
```

---

## 2. Navigation Graph

### Complete Navigation Graph

```kotlin
// ─── Route Definitions ─────────────────────────────────────

object Routes {
    // Auth
    const val SPLASH = "splash"
    const val LOGIN = "login"
    const val REGISTER = "register"
    const val ONBOARDING = "onboarding"

    // Main
    const val DASHBOARD = "dashboard"
    const val CHAT_LIST = "chat_list"
    const val CHAT_DETAIL = "chat/{conversationId}"
    const val AGENTS = "agents"
    const val AGENT_DETAIL = "agent/{agentId}"
    const val KNOWLEDGE = "knowledge"
    const val DOCUMENT_DETAIL = "document/{documentId}"

    // Settings
    const val SETTINGS = "settings"
    const val SETTINGS_PROFILE = "settings/profile"
    const val SETTINGS_SECURITY = "settings/security"
    const val SETTINGS_ABOUT = "settings/about"

    // Modal
    const val SEARCH = "search"
    const val PROFILE = "profile"
    const val CREATE_CONVERSATION = "create_conversation"
    const val EDIT_DOCUMENT = "edit_document/{documentId}"

    // Helpers
    fun chatDetail(conversationId: String) = "chat/$conversationId"
    fun agentDetail(agentId: String) = "agent/$agentId"
    fun documentDetail(documentId: String) = "document/$documentId"
    fun editDocument(documentId: String) = "edit_document/$documentId"
}

// ─── Graph Extensions ──────────────────────────────────────

fun NavGraphBuilder.authGraph(navController: NavHostController) {
    navigation(
        startDestination = Routes.SPLASH,
        route = "auth"
    ) {
        composable(Routes.SPLASH) {
            SplashScreen(
                onNavigateToLogin = {
                    navController.navigate(Routes.LOGIN) {
                        popUpTo("auth") { inclusive = true }
                    }
                },
                onNavigateToDashboard = {
                    navController.navigate(Routes.DASHBOARD) {
                        popUpTo("auth") { inclusive = true }
                    }
                }
            )
        }

        composable(Routes.LOGIN) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(Routes.DASHBOARD) {
                        popUpTo("auth") { inclusive = true }
                    }
                },
                onNavigateToRegister = {
                    navController.navigate(Routes.REGISTER)
                }
            )
        }

        composable(Routes.REGISTER) {
            RegisterScreen(
                onRegisterSuccess = {
                    navController.navigate(Routes.DASHBOARD) {
                        popUpTo("auth") { inclusive = true }
                    }
                },
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
    }
}

fun NavGraphBuilder.dashboardGraph(navController: NavHostController) {
    composable(
        route = Routes.DASHBOARD,
        enterTransition = { fadeIn(animationSpec = tween(300)) },
        exitTransition = { fadeOut(animationSpec = tween(300)) }
    ) {
        DashboardScreen(
            onNavigateToChat = { id ->
                navController.navigate(Routes.chatDetail(id))
            },
            onNavigateToDocument = { id ->
                navController.navigate(Routes.documentDetail(id))
            },
            onNavigateToAgent = { id ->
                navController.navigate(Routes.agentDetail(id))
            },
            onNavigateToSearch = {
                navController.navigate(Routes.SEARCH)
            }
        )
    }
}

fun NavGraphBuilder.chatGraph(navController: NavHostController) {
    composable(
        route = Routes.CHAT_LIST
    ) {
        ChatListScreen(
            onChatClick = { id ->
                navController.navigate(Routes.chatDetail(id))
            },
            onNewChat = {
                navController.navigate(Routes.CREATE_CONVERSATION)
            }
        )
    }

    composable(
        route = Routes.CHAT_DETAIL,
        arguments = listOf(
            navArgument("conversationId") { type = NavType.StringType }
        ),
        enterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Left) },
        exitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Left) },
        popEnterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Right) },
        popExitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Right) }
    ) { backStackEntry ->
        val conversationId = backStackEntry.arguments?.getString("conversationId") ?: return@composable

        ChatDetailScreen(
            conversationId = conversationId,
            onNavigateBack = { navController.popBackStack() },
            onNavigateToAgent = { agentId ->
                navController.navigate(Routes.agentDetail(agentId))
            },
            onNavigateToDocument = { docId ->
                navController.navigate(Routes.documentDetail(docId))
            }
        )
    }

    composable(
        route = Routes.CREATE_CONVERSATION,
        enterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Up) },
        popExitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Down) }
    ) {
        CreateConversationScreen(
            onConversationCreated = { id ->
                navController.popBackStack()
                navController.navigate(Routes.chatDetail(id))
            },
            onNavigateBack = { navController.popBackStack() }
        )
    }
}

fun NavGraphBuilder.agentsGraph(navController: NavHostController) {
    composable(Routes.AGENTS) {
        AgentsScreen(
            onAgentClick = { id ->
                navController.navigate(Routes.agentDetail(id))
            }
        )
    }

    composable(
        route = Routes.AGENT_DETAIL,
        arguments = listOf(
            navArgument("agentId") { type = NavType.StringType }
        )
    ) { backStackEntry ->
        val agentId = backStackEntry.arguments?.getString("agentId") ?: return@composable

        AgentDetailScreen(
            agentId = agentId,
            onNavigateBack = { navController.popBackStack() },
            onStartChat = { id ->
                navController.navigate(Routes.CREATE_CONVERSATION) {
                    launchSingleTop = true
                    popUpTo(Routes.AGENTS)
                }
            }
        )
    }
}

fun NavGraphBuilder.knowledgeGraph(navController: NavHostController) {
    composable(Routes.KNOWLEDGE) {
        KnowledgeScreen(
            onDocumentClick = { id ->
                navController.navigate(Routes.documentDetail(id))
            }
        )
    }

    composable(
        route = Routes.DOCUMENT_DETAIL,
        arguments = listOf(
            navArgument("documentId") { type = NavType.StringType }
        )
    ) { backStackEntry ->
        val documentId = backStackEntry.arguments?.getString("documentId") ?: return@composable

        DocumentDetailScreen(
            documentId = documentId,
            onNavigateBack = { navController.popBackStack() },
            onEditDocument = { id ->
                navController.navigate(Routes.editDocument(id))
            }
        )
    }
}

fun NavGraphBuilder.settingsGraph(navController: NavHostController) {
    composable(Routes.SETTINGS) {
        SettingsScreen(
            onNavigateToProfile = { navController.navigate(Routes.SETTINGS_PROFILE) },
            onNavigateToSecurity = { navController.navigate(Routes.SETTINGS_SECURITY) },
            onNavigateToAbout = { navController.navigate(Routes.SETTINGS_ABOUT) }
        )
    }

    composable(Routes.SETTINGS_PROFILE) {
        ProfileSettingsScreen(
            onNavigateBack = { navController.popBackStack() }
        )
    }

    composable(Routes.SETTINGS_SECURITY) {
        SecuritySettingsScreen(
            onNavigateBack = { navController.popBackStack() }
        )
    }

    composable(Routes.SETTINGS_ABOUT) {
        AboutScreen(
            onNavigateBack = { navController.popBackStack() }
        )
    }
}

fun NavGraphBuilder.profileGraph(navController: NavHostController) {
    dialog(
        route = Routes.PROFILE,
        enterTransition = { fadeIn() },
        exitTransition = { fadeOut() }
    ) {
        ProfileDialog(
            onDismiss = { navController.popBackStack() }
        )
    }
}

fun NavGraphBuilder.searchGraph(navController: NavHostController) {
    composable(
        route = Routes.SEARCH,
        enterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Up) },
        popExitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Down) }
    ) {
        SearchScreen(
            onResultClick = { type, id ->
                when (type) {
                    "conversation" -> navController.navigate(Routes.chatDetail(id))
                    "document" -> navController.navigate(Routes.documentDetail(id))
                    "agent" -> navController.navigate(Routes.agentDetail(id))
                }
            },
            onNavigateBack = { navController.popBackStack() }
        )
    }
}
```

---

## 3. Route Definitions

### Type-Safe Route Definitions

```kotlin
// ─── Sealed Class Routes ───────────────────────────────────

sealed class Screen(val route: String) {
    data object Splash : Screen("splash")
    data object Login : Screen("login")
    data object Register : Screen("register")
    data object Dashboard : Screen("dashboard")
    data object ChatList : Screen("chat_list")

    data class ChatDetail(val conversationId: String) : Screen("chat/{conversationId}") {
        fun createRoute(conversationId: String) = "chat/$conversationId"
    }

    data object Agents : Screen("agents")

    data class AgentDetail(val agentId: String) : Screen("agent/{agentId}") {
        fun createRoute(agentId: String) = "agent/$agentId"
    }

    data object Knowledge : Screen("knowledge")

    data class DocumentDetail(val documentId: String) : Screen("document/{documentId}") {
        fun createRoute(documentId: String) = "document/$documentId"
    }

    data object Settings : Screen("settings")
    data object Search : Screen("search")
    data object Profile : Screen("profile")
    data object CreateConversation : Screen("create_conversation")
}

// ─── Usage ─────────────────────────────────────────────────

// Navigation
navController.navigate(Screen.ChatDetail.createConversation("conv-123"))

// In NavHost
composable(
    route = Screen.ChatDetail.route,
    arguments = listOf(
        navArgument("conversationId") { type = NavType.StringType }
    )
) { backStackEntry ->
    val conversationId = backStackEntry.arguments?.getString("conversationId")
    ChatDetailScreen(conversationId = conversationId!!)
}
```

---

## 4. Route Arguments

### Argument Types

```kotlin
// ─── Path Arguments ────────────────────────────────────────

// String argument
navArgument("conversationId") {
    type = NavType.StringType
    nullable = false
}

// Int argument
navArgument("page") {
    type = NavType.IntType
    defaultValue = 0
}

// Long argument
navArgument("timestamp") {
    type = NavType.LongType
    defaultValue = System.currentTimeMillis()
}

// Boolean argument
navArgument("isFromNotification") {
    type = NavType.BoolType
    defaultValue = false
}

// Float argument
navArgument("latitude") {
    type = NavType.FloatType
    defaultValue = 0f
}

// ─── Query Parameters ──────────────────────────────────────

// Passing query params with navigation
navController.navigate("search?q=kotlin&type=conversation") {
    // or using buildUpon
    val route = Uri.Builder()
        .appendPath("search")
        .appendQueryParameter("q", "kotlin")
        .appendQueryParameter("type", "conversation")
        .build().toString()
    navController.navigate(route)
}

// Reading query params
composable("search?q={query}&type={type}") { backStackEntry ->
    val query = backStackEntry.arguments?.getString("q") ?: ""
    val type = backStackEntry.arguments?.getString("type") ?: "all"
    SearchScreen(query = query, type = type)
}

// ─── Complex Arguments with Serialization ──────────────────

@Serializable
data class ChatNavArgs(
    val conversationId: String,
    val title: String? = null,
    val agentId: String? = null
)

// Using kotlinx.serialization with navigation
composable<ChatNavArgs> { backStackEntry ->
    val args = backStackEntry.toRoute<ChatNavArgs>()
    ChatDetailScreen(
        conversationId = args.conversationId,
        title = args.title,
        agentId = args.agentId
    )
}

// Navigation with typed args
navController.navigate(
    ChatNavArgs(
        conversationId = "conv-123",
        title = "My Chat"
    )
)
```

### Argument Parsing Table

| Type | NavType | Default | Nullable | Example |
|------|---------|---------|----------|---------|
| String | `NavType.StringType` | Yes | Yes | `"conv-123"` |
| Int | `NavType.IntType` | Yes | No | `42` |
| Long | `NavType.LongType` | Yes | No | `123456789L` |
| Float | `NavType.FloatType` | Yes | No | `3.14f` |
| Boolean | `NavType.BoolType` | Yes | No | `true` |
| String[] | `NavType.StringArrayType` | Yes | No | `["a","b"]` |
| Int[] | `NavType.IntArrayType` | Yes | No | `[1,2,3]` |

---

## 5. Screen Transitions

### Transition Types

```kotlin
// ─── Fade Transition ───────────────────────────────────────

composable(
    route = Routes.DASHBOARD,
    enterTransition = { fadeIn(animationSpec = tween(300)) },
    exitTransition = { fadeOut(animationSpec = tween(300)) },
    popEnterTransition = { fadeIn(animationSpec = tween(300)) },
    popExitTransition = { fadeOut(animationSpec = tween(300)) }
) {
    DashboardScreen()
}

// ─── Slide Transitions ─────────────────────────────────────

composable(
    route = Routes.CHAT_DETAIL,
    enterTransition = {
        slideIntoContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Left,
            animationSpec = tween(300)
        )
    },
    exitTransition = {
        slideOutOfContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Left,
            animationSpec = tween(300)
        )
    },
    popEnterTransition = {
        slideIntoContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Right,
            animationSpec = tween(300)
        )
    },
    popExitTransition = {
        slideOutOfContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Right,
            animationSpec = tween(300)
        )
    }
) {
    ChatDetailScreen()
}

// ─── Scale Transition ──────────────────────────────────────

composable(
    route = Routes.PROFILE,
    enterTransition = {
        scaleIn(
            initialScale = 0.8f,
            animationSpec = tween(300)
        ) + fadeIn()
    },
    exitTransition = {
        scaleOut(
            targetScale = 0.8f,
            animationSpec = tween(300)
        ) + fadeOut()
    }
) {
    ProfileDialog()
}

// ─── Shared Element Transition ─────────────────────────────

composable(
    route = Routes.AGENT_DETAIL,
    enterTransition = {
        fadeIn() + slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Up)
    },
    exitTransition = {
        fadeOut() + slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Down)
    }
) { backStackEntry ->
    AgentDetailScreen(
        agentId = backStackEntry.arguments?.getString("agentId") ?: ""
    )
}

// ─── Custom Transition Builder ─────────────────────────────

fun NavGraphBuilder.customTransition(
    route: String,
    content: @Composable AnimatedContentScope.(NavBackStackEntry) -> Unit
) {
    composable(
        route = route,
        enterTransition = {
            fadeIn(animationSpec = tween(300)) + slideInHorizontally(
                initialOffsetX = { it / 3 },
                animationSpec = tween(300)
            )
        },
        exitTransition = {
            fadeOut(animationSpec = tween(300)) + slideOutHorizontally(
                targetOffsetX = { -it / 3 },
                animationSpec = tween(300)
            )
        },
        popEnterTransition = {
            fadeIn(animationSpec = tween(300)) + slideInHorizontally(
                initialOffsetX = { -it / 3 },
                animationSpec = tween(300)
            )
        },
        popExitTransition = {
            fadeOut(animationSpec = tween(300)) + slideOutHorizontally(
                targetOffsetX = { it / 3 },
                animationSpec = tween(300)
            )
        },
        content = content
    )
}

// ─── Animated NavHost ─────────────────────────────────────

@Composable
fun AnimatedNexusNavHost(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    AnimatedNavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD,
        modifier = modifier,
        enterTransition = { fadeIn(tween(200)) },
        exitTransition = { fadeOut(tween(200)) },
        popEnterTransition = { fadeIn(tween(200)) },
        popExitTransition = { fadeOut(tween(200)) }
    ) {
        // Navigation graph...
    }
}
```

---

## 6. Bottom Navigation

### Bottom Navigation Implementation

```kotlin
// ─── Bottom Nav Items ──────────────────────────────────────

enum class BottomNavItem(
    val route: String,
    val label: String,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector
) {
    DASHBOARD(
        route = Routes.DASHBOARD,
        label = "Dashboard",
        selectedIcon = Icons.Filled.Dashboard,
        unselectedIcon = Icons.Outlined.Dashboard
    ),
    CHAT(
        route = Routes.CHAT_LIST,
        label = "Chat",
        selectedIcon = Icons.Filled.Chat,
        unselectedIcon = Icons.Outlined.Chat
    ),
    AGENTS(
        route = Routes.AGENTS,
        label = "Agents",
        selectedIcon = Icons.Filled.SmartToy,
        unselectedIcon = Icons.Outlined.SmartToy
    ),
    KNOWLEDGE(
        route = Routes.KNOWLEDGE,
        label = "Knowledge",
        selectedIcon = Icons.Filled.LibraryBooks,
        unselectedIcon = Icons.Outlined.LibraryBooks
    ),
    SETTINGS(
        route = Routes.SETTINGS,
        label = "More",
        selectedIcon = Icons.Filled.MoreHoriz,
        unselectedIcon = Icons.Outlined.MoreHoriz
    )
}

val bottomBarRoutes = BottomNavItem.entries.map { it.route }

// ─── Bottom Bar Component ──────────────────────────────────

@Composable
fun NexusBottomBar(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    NavigationBar(
        modifier = modifier,
        containerColor = MaterialTheme.colorScheme.surfaceContainer,
        contentColor = MaterialTheme.colorScheme.onSurface
    ) {
        BottomNavItem.entries.forEach { item ->
            val isSelected = currentRoute == item.route
            val unreadCount = getUnreadCount(item.route) // From ViewModel

            NavigationBarItem(
                selected = isSelected,
                onClick = {
                    if (currentRoute != item.route) {
                        navController.navigate(item.route) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                },
                icon = {
                    BadgedBox(
                        badge = {
                            if (unreadCount > 0) {
                                Badge {
                                    Text(text = if (unreadCount > 99) "99+" else unreadCount.toString())
                                }
                            }
                        }
                    ) {
                        Icon(
                            imageVector = if (isSelected) item.selectedIcon else item.unselectedIcon,
                            contentDescription = item.label
                        )
                    }
                },
                label = {
                    Text(
                        text = item.label,
                        style = MaterialTheme.typography.labelSmall
                    )
                },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = MaterialTheme.colorScheme.onPrimaryContainer,
                    selectedTextColor = MaterialTheme.colorScheme.onPrimaryContainer,
                    indicatorColor = MaterialTheme.colorScheme.primaryContainer
                )
            )
        }
    }
}

// ─── Navigation Bar with Scroll Behavior ───────────────────

@Composable
fun ScrollableBottomBar(
    navController: NavHostController,
    scrollBehavior: TopAppBarScrollBehavior,
    modifier: Modifier = Modifier
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    // Hide bottom bar when scrolling
    val isVisible by remember {
        derivedStateOf { scrollBehavior.state.heightOffset == 0f }
    }

    AnimatedVisibility(
        visible = isVisible,
        enter = slideInVertically(initialOffsetY = { it }),
        exit = slideOutVertically(targetOffsetY = { it })
    ) {
        NavigationBar(modifier = modifier) {
            // Items...
        }
    }
}
```

---

## 7. Navigation Bar Items

### Complete Navigation Bar

```
┌─────────────────────────────────────────────────────────┐
│                    Bottom Navigation                     │
│                                                         │
│  ┌──────────┬──────────┬──────────┬──────────┬───────┐ │
│  │Dashboard │   Chat   │  Agents  │Knowledge │ More  │ │
│  │  ┌────┐  │  ┌────┐  │  ┌────┐  │  ┌────┐  │ ┌───┐│ │
│  │  │ 📊 │  │  │ 💬 │  │  │ 🤖 │  │  │ 📚 │  │ │☰  ││ │
│  │  └────┘  │  └────┘  │  └────┘  │  └────┘  │ └───┘│ │
│  │Dashboard │   Chat   │  Agents  │Knowledge │  More │ │
│  │   ● ●●   │          │          │          │       │ │
│  └──────────┴──────────┴──────────┴──────────┴───────┘ │
│                                                         │
│  ● = selected indicator    ●● = badge (unread count)   │
└─────────────────────────────────────────────────────────┘
```

### Unread Count Management

```kotlin
@HiltViewModel
class NavigationViewModel @Inject constructor(
    private val conversationRepository: ConversationRepository,
    private val documentRepository: DocumentRepository
) : ViewModel() {

    data class NavigationState(
        val unreadConversations: Int = 0,
        val pendingDocuments: Int = 0,
        val pendingSync: Int = 0
    )

    private val _state = MutableStateFlow(NavigationState())
    val state: StateFlow<NavigationState> = _state.asStateFlow()

    init {
        observeUnreadCounts()
    }

    private fun observeUnreadCounts() {
        viewModelScope.launch {
            combine(
                conversationRepository.getUnreadCount(),
                documentRepository.getPendingCount(),
                syncManager.getPendingCount()
            ) { conversations, documents, sync ->
                NavigationState(
                    unreadConversations = conversations,
                    pendingDocuments = documents,
                    pendingSync = sync
                )
            }.collect { state ->
                _state.value = state
            }
        }
    }
}

@Composable
fun NexusBottomBar(
    navController: NavHostController,
    viewModel: NavigationViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    NavigationBar {
        item(
            selected = currentRoute == Routes.DASHBOARD,
            onClick = { /* navigate */ },
            icon = {
                BadgedBox(
                    badge = {
                        if (state.pendingSync > 0) {
                            Badge { Text("${state.pendingSync}") }
                        }
                    }
                ) {
                    Icon(Icons.Filled.Dashboard, contentDescription = "Dashboard")
                }
            },
            label = { Text("Dashboard") }
        )

        item(
            selected = currentRoute == Routes.CHAT_LIST,
            onClick = { /* navigate */ },
            icon = {
                BadgedBox(
                    badge = {
                        if (state.unreadConversations > 0) {
                            Badge { Text("${state.unreadConversations}") }
                        }
                    }
                ) {
                    Icon(Icons.Filled.Chat, contentDescription = "Chat")
                }
            },
            label = { Text("Chat") }
        )

        item(
            selected = currentRoute == Routes.AGENTS,
            onClick = { /* navigate */ },
            icon = { Icon(Icons.Filled.SmartToy, contentDescription = "Agents") },
            label = { Text("Agents") }
        )

        item(
            selected = currentRoute == Routes.KNOWLEDGE,
            onClick = { /* navigate */ },
            icon = {
                BadgedBox(
                    badge = {
                        if (state.pendingDocuments > 0) {
                            Badge { Text("${state.pendingDocuments}") }
                        }
                    }
                ) {
                    Icon(Icons.Filled.LibraryBooks, contentDescription = "Knowledge")
                }
            },
            label = { Text("Knowledge") }
        )

        item(
            selected = currentRoute == Routes.SETTINGS,
            onClick = { /* navigate */ },
            icon = { Icon(Icons.Filled.MoreHoriz, contentDescription = "More") },
            label = { Text("More") }
        )
    }
}
```

---

## 8. Top App Bar

### Top App Bar Implementation

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusTopBar(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    when (currentRoute) {
        Routes.DASHBOARD -> DashboardTopBar(modifier)
        Routes.CHAT_LIST -> ChatListTopBar(modifier)
        Routes.AGENTS -> AgentsTopBar(modifier)
        Routes.KNOWLEDGE -> KnowledgeTopBar(modifier)
        Routes.SETTINGS -> SettingsTopBar(modifier)
        Routes.CHAT_DETAIL -> ChatDetailTopBar(
            navController = navController,
            modifier = modifier
        )
        Routes.AGENT_DETAIL -> AgentDetailTopBar(
            navController = navController,
            modifier = modifier
        )
        else -> null // No top bar for other screens
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardTopBar(modifier: Modifier = Modifier) {
    TopAppBar(
        title = {
            Text(
                text = "Nexus AI",
                style = MaterialTheme.typography.headlineSmall
            )
        },
        actions = {
            IconButton(onClick = { /* navigate to search */ }) {
                Icon(Icons.Filled.Search, contentDescription = "Search")
            }
            IconButton(onClick = { /* navigate to profile */ }) {
                Icon(Icons.Filled.AccountCircle, contentDescription = "Profile")
            }
        },
        colors = TopAppBarDefaults.topAppBarColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        modifier = modifier
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatDetailTopBar(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val conversationTitle = navBackStackEntry?.arguments?.getString("title") ?: "Chat"

    TopAppBar(
        title = {
            Column {
                Text(
                    text = conversationTitle,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                // Online indicator
                Text(
                    text = "Online",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        },
        navigationIcon = {
            IconButton(onClick = { navController.popBackStack() }) {
                Icon(Icons.Filled.ArrowBack, contentDescription = "Back")
            }
        },
        actions = {
            IconButton(onClick = { /* Call */ }) {
                Icon(Icons.Filled.Phone, contentDescription = "Call")
            }
            IconButton(onClick = { /* Video */ }) {
                Icon(Icons.Filled.Videocam, contentDescription = "Video")
            }
            IconButton(onClick = { /* More options */ }) {
                Icon(Icons.Filled.MoreVert, contentDescription = "More options")
            }
        },
        colors = TopAppBarDefaults.topAppBarColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        modifier = modifier
    )
}

// ─── Scroll-aware Top Bar ──────────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScrollAwareTopBar(
    title: String,
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    val scrollBehavior = TopAppBarDefaults.exitUntilCollapsedScrollBehavior()

    Scaffold(
        topBar = {
            LargeTopAppBar(
                title = { Text(title) },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                scrollBehavior = scrollBehavior,
                colors = TopAppBarDefaults.largeTopAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { innerPadding ->
        LazyColumn(
            contentPadding = innerPadding,
            modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection)
        ) {
            // Content
        }
    }
}
```

---

## 9. Navigation Drawer

### Navigation Drawer Implementation

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusNavigationDrawer(
    navController: NavHostController,
    drawerState: DrawerState = rememberDrawerState(initialValue = DrawerValue.Closed),
    content: @Composable () -> Unit
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    ModalNavigationDrawer(
        drawerState = drawerState,
        gesturesEnabled = currentRoute in drawerEnabledRoutes,
        drawerContent = {
            ModalDrawerSheet {
                // Drawer Header
                DrawerHeader(
                    userName = "Ali",
                    userEmail = "ali@example.com",
                    avatarUrl = null
                )

                HorizontalDivider()

                // Drawer Items
                NavigationDrawerItem(
                    label = { Text("Dashboard") },
                    icon = { Icon(Icons.Filled.Dashboard, contentDescription = null) },
                    selected = currentRoute == Routes.DASHBOARD,
                    onClick = {
                        navController.navigate(Routes.DASHBOARD) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                        CoroutineScope(Dispatchers.IO).launch { drawerState.close() }
                    }
                )

                NavigationDrawerItem(
                    label = { Text("Chat") },
                    icon = { Icon(Icons.Filled.Chat, contentDescription = null) },
                    selected = currentRoute == Routes.CHAT_LIST,
                    onClick = {
                        navController.navigate(Routes.CHAT_LIST) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                        CoroutineScope(Dispatchers.IO).launch { drawerState.close() }
                    }
                )

                NavigationDrawerItem(
                    label = { Text("Agents") },
                    icon = { Icon(Icons.Filled.SmartToy, contentDescription = null) },
                    selected = currentRoute == Routes.AGENTS,
                    onClick = {
                        navController.navigate(Routes.AGENTS) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                        CoroutineScope(Dispatchers.IO).launch { drawerState.close() }
                    }
                )

                HorizontalDivider()

                NavigationDrawerItem(
                    label = { Text("Settings") },
                    icon = { Icon(Icons.Filled.Settings, contentDescription = null) },
                    selected = currentRoute == Routes.SETTINGS,
                    onClick = {
                        navController.navigate(Routes.SETTINGS)
                        CoroutineScope(Dispatchers.IO).launch { drawerState.close() }
                    }
                )

                NavigationDrawerItem(
                    label = { Text("About") },
                    icon = { Icon(Icons.Filled.Info, contentDescription = null) },
                    selected = currentRoute == Routes.SETTINGS_ABOUT,
                    onClick = {
                        navController.navigate(Routes.SETTINGS_ABOUT)
                        CoroutineScope(Dispatchers.IO).launch { drawerState.close() }
                    }
                )
            }
        }
    ) {
        content()
    }
}

@Composable
fun DrawerHeader(
    userName: String,
    userEmail: String,
    avatarUrl: String?,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp)
            .statusBarsPadding(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // Avatar
        if (avatarUrl != null) {
            AsyncImage(
                model = avatarUrl,
                contentDescription = "Profile picture",
                modifier = Modifier
                    .size(72.dp)
                    .clip(CircleShape),
                contentScale = ContentScale.Crop
            )
        } else {
            Icon(
                Icons.Filled.AccountCircle,
                contentDescription = "Profile picture",
                modifier = Modifier.size(72.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        Text(
            text = userName,
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.onSurface
        )

        Text(
            text = userEmail,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

val drawerEnabledRoutes = listOf(
    Routes.DASHBOARD,
    Routes.CHAT_LIST,
    Routes.AGENTS,
    Routes.KNOWLEDGE,
    Routes.SETTINGS
)
```

---

## 10. Navigation Actions

### Navigation Action Patterns

```kotlin
// ─── Basic Navigation ──────────────────────────────────────

// Navigate to a route
navController.navigate("chat/conv-123")

// Navigate with popUpTo (singleTop behavior)
navController.navigate("dashboard") {
    popUpTo(navController.graph.findStartDestination().id) {
        saveState = true
    }
    launchSingleTop = true
    restoreState = true
}

// Navigate and clear back stack
navController.navigate("dashboard") {
    popUpTo(0) { inclusive = true }
}

// Navigate and pop to specific route
navController.navigate("chat/conv-123") {
    popUpTo("dashboard")
}

// ─── Pop Back Stack ────────────────────────────────────────

// Simple pop
navController.popBackStack()

// Pop to specific route
navController.popUpTo("dashboard") { inclusive = false }

// Pop inclusive (removes target from stack)
navController.popUpTo("dashboard") { inclusive = true }

// ─── Navigate Up ───────────────────────────────────────────

navController.navigateUp()

// ─── Check Navigation State ────────────────────────────────

val canGoBack = navController.previousBackStackEntry != null
val currentRoute = navController.currentBackStackEntry?.destination?.route

// ─── Navigation with Results ───────────────────────────────

// In caller
val result = navController.currentBackStackEntry
    ?.savedStateHandle
    ?.getStateFlow<String>("selected_item", "")
    ?.collectAsState()

// Set result before popping
navController.previousBackStackEntry
    ?.savedStateHandle
    ?.set("selected_item", "item-id")

navController.popBackStack()

// ─── Navigation Action Sealed Class ────────────────────────

sealed class NavigationAction {
    data class Navigate(val route: String) : NavigationAction()
    data class NavigateWithArgs(val route: String, val args: Map<String, String>) : NavigationAction()
    data object NavigateBack : NavigationAction()
    data class PopTo(val route: String, val inclusive: Boolean = false) : NavigationAction()
    data class NavigateAndClear(val route: String) : NavigationAction()
}

fun executeNavigationAction(
    navController: NavHostController,
    action: NavigationAction
) {
    when (action) {
        is NavigationAction.Navigate -> {
            navController.navigate(action.route)
        }
        is NavigationAction.NavigateWithArgs -> {
            val route = buildRoute(action.route, action.args)
            navController.navigate(route)
        }
        is NavigationAction.NavigateBack -> {
            navController.popBackStack()
        }
        is NavigationAction.PopTo -> {
            navController.popUpTo(action.route) { inclusive = action.inclusive }
        }
        is NavigationAction.NavigateAndClear -> {
            navController.navigate(action.route) {
                popUpTo(0) { inclusive = true }
            }
        }
    }
}

fun buildRoute(baseRoute: String, args: Map<String, String>): String {
    var route = baseRoute
    args.forEach { (key, value) ->
        route = route.replace("{$key}", value)
    }
    return route
}
```

---

## 11. Deep Links

### Deep Link Configuration

```kotlin
// ─── AndroidManifest.xml ───────────────────────────────────

<activity
    android:name=".MainActivity"
    android:launchMode="singleTop"
    android:exported="true"
    android:theme="@style/Theme.NexusAI">

    <!-- App links -->
    <intent-filter android:autoVerify="true">
        <action android:name="android.intent.action.VIEW" />
        <category android:name="android.intent.category.DEFAULT" />
        <category android:name="android.intent.category.BROWSABLE" />
        <data
            android:scheme="https"
            android:host="nexus-ai.com"
            android:pathPrefix="/app" />
    </intent-filter>

    <!-- Custom scheme -->
    <intent-filter>
        <action android:name="android.intent.action.VIEW" />
        <category android:name="android.intent.category.DEFAULT" />
        <category android:name="android.intent.category.BROWSABLE" />
        <data
            android:scheme="nexusai"
            android:host="open" />
    </intent-filter>
</activity>

// ─── Navigation Deep Links ─────────────────────────────────

fun NavGraphBuilder.withDeepLinks() {
    composable(
        route = Routes.CHAT_DETAIL,
        arguments = listOf(
            navArgument("conversationId") { type = NavType.StringType }
        ),
        deepLinks = listOf(
            navDeepLink {
                uriPattern = "https://nexus-ai.com/app/chat/{conversationId}"
                action = Intent.ACTION_VIEW
            },
            navDeepLink {
                uriPattern = "nexusai://open/chat/{conversationId}"
            }
        )
    ) { backStackEntry ->
        val conversationId = backStackEntry.arguments?.getString("conversationId")
        ChatDetailScreen(conversationId = conversationId ?: "")
    }

    composable(
        route = Routes.AGENT_DETAIL,
        deepLinks = listOf(
            navDeepLink {
                uriPattern = "https://nexus-ai.com/app/agent/{agentId}"
            },
            navDeepLink {
                uriPattern = "nexusai://open/agent/{agentId}"
            }
        )
    ) { backStackEntry ->
        val agentId = backStackEntry.arguments?.getString("agentId")
        AgentDetailScreen(agentId = agentId ?: "")
    }

    composable(
        route = Routes.KNOWLEDGE,
        deepLinks = listOf(
            navDeepLink {
                uriPattern = "https://nexus-ai.com/app/knowledge"
            },
            navDeepLink {
                uriPattern = "nexusai://open/knowledge"
            }
        )
    ) {
        KnowledgeScreen()
    }

    composable(
        route = Routes.DOCUMENT_DETAIL,
        deepLinks = listOf(
            navDeepLink {
                uriPattern = "https://nexus-ai.com/app/document/{documentId}"
            }
        )
    ) { backStackEntry ->
        val documentId = backStackEntry.arguments?.getString("documentId")
        DocumentDetailScreen(documentId = documentId ?: "")
    }
}
```

### Deep Link Diagram

```
┌─────────────────────────────────────────────────────────┐
│                  Deep Link Flow                          │
│                                                         │
│  ┌──────────┐     ┌──────────┐     ┌──────────────┐   │
│  │ Browser/ │     │Android   │     │  Navigation  │   │
│  │ Sharing  │────▶│Manifest  │────▶│  Controller  │   │
│  │          │     │(Intent   │     │              │   │
│  └──────────┘     │Filter)   │     └──────┬───────┘   │
│                   └──────────┘            │            │
│                                    ┌──────▼───────┐    │
│                                    │  Route       │    │
│                                    │  Matching    │    │
│                                    └──────┬───────┘    │
│                                    ┌──────┼──────────┐ │
│                                    ▼      ▼          ▼ │
│                              ┌──────┐ ┌──────┐ ┌──────┐│
│                              │ Chat │ │Agent │ │ Doc  ││
│                              │Detail│ │Detail│ │Detail││
│                              └──────┘ └──────┘ └──────┘│
└─────────────────────────────────────────────────────────┘
```

---

## 12. Deep Link Handling

### Deep Link Handler

```kotlin
class DeepLinkHandler @Inject constructor(
    private val signatureVerifier: SignatureVerifier
) {
    sealed class DeepLinkResult {
        data class Valid(val route: String) : DeepLinkResult()
        data object Invalid : DeepLinkResult()
        data object Expired : DeepLinkResult()
    }

    fun handleIntent(intent: Intent): DeepLinkResult {
        val data = intent.data ?: return DeepLinkResult.Invalid

        return when (data.scheme) {
            "https" -> handleHttpsLink(data)
            "nexusai" -> handleCustomScheme(data)
            else -> DeepLinkResult.Invalid
        }
    }

    private fun handleHttpsLink(uri: Uri): DeepLinkResult {
        if (uri.host != "nexus-ai.com") return DeepLinkResult.Invalid

        val pathSegments = uri.pathSegments
        if (pathSegments.firstOrNull() != "app") return DeepLinkResult.Invalid

        val route = when (pathSegments.getOrNull(1)) {
            "chat" -> {
                val id = pathSegments.getOrNull(2) ?: return DeepLinkResult.Invalid
                Routes.chatDetail(id)
            }
            "agent" -> {
                val id = pathSegments.getOrNull(2) ?: return DeepLinkResult.Invalid
                Routes.agentDetail(id)
            }
            "document" -> {
                val id = pathSegments.getOrNull(2) ?: return DeepLinkResult.Invalid
                Routes.documentDetail(id)
            }
            "knowledge" -> Routes.KNOWLEDGE
            "dashboard" -> Routes.DASHBOARD
            else -> DeepLinkResult.Invalid
        }

        // Verify signature if present
        val sig = uri.getQueryParameter("sig")
        val ts = uri.getQueryParameter("ts")
        if (sig != null && ts != null) {
            if (!verifyLinkIntegrity(uri, sig, ts)) {
                return DeepLinkResult.Invalid
            }
        }

        return if (route is String) DeepLinkResult.Valid(route) else DeepLinkResult.Invalid
    }

    private fun handleCustomScheme(uri: Uri): DeepLinkResult {
        if (uri.host != "open") return DeepLinkResult.Invalid

        val route = when (uri.pathSegments.firstOrNull()) {
            "chat" -> {
                val id = uri.pathSegments.getOrNull(1) ?: return DeepLinkResult.Invalid
                Routes.chatDetail(id)
            }
            "agent" -> {
                val id = uri.pathSegments.getOrNull(1) ?: return DeepLinkResult.Invalid
                Routes.agentDetail(id)
            }
            "knowledge" -> Routes.KNOWLEDGE
            "settings" -> Routes.SETTINGS
            else -> DeepLinkResult.Invalid
        }

        return DeepLinkResult.Valid(route)
    }

    private fun verifyLinkIntegrity(uri: Uri, signature: String, timestamp: String): Boolean {
        val ts = timestamp.toLongOrNull() ?: return false

        // Check freshness (5 minutes)
        if (System.currentTimeMillis() - ts > 5 * 60 * 1000) return false

        // Verify HMAC signature
        val payload = "${uri.host}${uri.path}:$timestamp"
        return signatureVerifier.verify(payload, signature)
    }
}

// ─── Usage in Activity ─────────────────────────────────────

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var deepLinkHandler: DeepLinkHandler

    private lateinit var navController: NavHostController

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        setContent {
            NexusTheme {
                navController = rememberNavController()

                // Handle initial deep link
                LaunchedEffect(intent) {
                    handleDeepLink(intent)
                }

                Scaffold { innerPadding ->
                    NexusNavHost(
                        navController = navController,
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleDeepLink(intent)
    }

    private fun handleDeepLink(intent: Intent?) {
        if (intent == null) return

        when (val result = deepLinkHandler.handleIntent(intent)) {
            is DeepLinkResult.Valid -> {
                navController.navigate(result.route) {
                    popUpTo(navController.graph.findStartDestination().id) {
                        saveState = true
                    }
                    launchSingleTop = true
                    restoreState = true
                }
            }
            is DeepLinkResult.Invalid -> {
                // Show error or do nothing
            }
            is DeepLinkResult.Expired -> {
                // Show expired link message
            }
        }
    }
}
```

---

## 13. Back Stack Management

### Back Stack Patterns

```kotlin
// ─── Save State and Restore ────────────────────────────────

navController.navigate(route) {
    popUpTo(navController.graph.findStartDestination().id) {
        saveState = true  // Save state of popped destinations
    }
    launchSingleTop = true  // Don't create multiple copies
    restoreState = true  // Restore previously saved state
}

// ─── Clear Back Stack ──────────────────────────────────────

// Navigate to login, clear entire stack
navController.navigate(Routes.LOGIN) {
    popUpTo(0) { inclusive = true }
}

// Navigate to dashboard, keep login in stack
navController.navigate(Routes.DASHBOARD) {
    popUpTo(Routes.LOGIN) { inclusive = false }
}

// ─── Pop Multiple Entries ──────────────────────────────────

// Pop to root
navController.popBackStack(Routes.DASHBOARD, inclusive = false)

// Pop until specific route
navController.popUpTo("dashboard") {
    inclusive = false
    saveState = true
}

// ─── Back Stack Count ──────────────────────────────────────

val backStackCount = navController.backQueue.size

// ─── Listen to Back Stack Changes ──────────────────────────

@Composable
fun BackStackListener(navController: NavHostController) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()

    LaunchedEffect(navBackStackEntry) {
        // Called when back stack changes
        val currentRoute = navBackStackEntry?.destination?.route
        // Track analytics, update UI, etc.
    }
}

// ─── Custom Back Handler ───────────────────────────────────

@Composable
fun ScreenWithCustomBack(
    navController: NavHostController,
    hasUnsavedChanges: Boolean
) {
    val backHandler = remember { BackHandler() }

    BackHandler(enabled = hasUnsavedChanges) {
        // Show confirmation dialog
        backHandler.showConfirmation(
            title = "Unsaved Changes",
            message = "You have unsaved changes. Are you sure you want to go back?",
            onConfirm = { navController.popBackStack() },
            onCancel = { /* Do nothing */ }
        )
    }

    // Screen content...
}

// ─── Back Stack Entry Properties ───────────────────────────

@Composable
fun BackStackInfo(navController: NavHostController) {
    val currentEntry by navController.currentBackStackEntryAsState()
    val previousEntry by navController.previousBackStackEntryAsState()

    Column {
        Text("Current: ${currentEntry?.destination?.route}")
        Text("Previous: ${previousEntry?.destination?.route}")
        Text("Back Stack Size: ${navController.backQueue.size}")
        Text("Can Go Back: ${navController.previousBackStackEntry != null}")
    }
}
```

### Back Stack Diagram

```
┌─────────────────────────────────────────────────────────┐
│                  Back Stack Example                      │
│                                                         │
│  After: Dashboard → Chat List → Chat Detail → Settings  │
│                                                         │
│  Back Stack:                                            │
│  ┌──────────┐                                          │
│  │ Settings │  ← Current (top)                         │
│  ├──────────┤                                          │
│  │Chat Detail│                                          │
│  ├──────────┤                                          │
│  │Chat List │                                          │
│  ├──────────┤                                          │
│  │Dashboard │  ← Start destination                     │
│  └──────────┘                                          │
│                                                         │
│  popBackStack():                                        │
│  Before: [Dashboard, ChatList, ChatDetail, Settings]   │
│  After:  [Dashboard, ChatList, ChatDetail]             │
│                                                         │
│  popBackStack("dashboard", inclusive=false):            │
│  Before: [Dashboard, ChatList, ChatDetail, Settings]   │
│  After:  [Dashboard]                                   │
│                                                         │
│  navigate("new_screen") { popUpTo(0) { inclusive } }:   │
│  Before: [Dashboard, ChatList, ChatDetail, Settings]   │
│  After:  [NewScreen]                                   │
└─────────────────────────────────────────────────────────┘
```

---

## 14. Navigation State Preservation

### State Preservation with Navigation

```kotlin
// ─── Save and Restore State ────────────────────────────────

@Composable
fun NexusNavHost(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD
    ) {
        // Screens with saveState/restoreState
        composable(
            route = Routes.DASHBOARD,
            enterTransition = { fadeIn() },
            exitTransition = { fadeOut() }
        ) {
            DashboardScreen()
        }

        composable(
            route = Routes.CHAT_LIST
        ) {
            ChatListScreen()
        }

        composable(
            route = Routes.AGENTS
        ) {
            AgentsScreen()
        }
    }
}

// Bottom bar navigation with state preservation
navController.navigate(item.route) {
    popUpTo(navController.graph.findStartDestination().id) {
        saveState = true  // Save state before navigating away
    }
    launchSingleTop = true  // Prevent duplicate destinations
    restoreState = true  // Restore state if previously visited
}

// ─── ViewModel State Preservation ──────────────────────────

@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val savedStateHandle: SavedStateHandle
) : ViewModel() {

    // Automatically saved/restored across navigation
    val scrollPosition: StateFlow<Int> = savedStateHandle.getStateFlow("scroll", 0)
    val selectedTab: StateFlow<Int> = savedStateHandle.getStateFlow("tab", 0)
    val searchQuery: StateFlow<String> = savedStateHandle.getStateFlow("query", "")

    fun saveScroll(position: Int) {
        savedStateHandle["scroll"] = position
    }

    fun selectTab(index: Int) {
        savedStateHandle["tab"] = index
    }
}

// ─── State Preservation Decision ───────────────────────────

/*
┌───────────────────────────────────────────────┐
│         State Preservation Strategy           │
│                                               │
│  Component State:                             │
│  - Use remember { } for transient state       │
│  - Use rememberSaveable { } for persistent    │
│                                               │
│  ViewModel State:                             │
│  - Automatic preservation across config change │
│  - Use SavedStateHandle for process death     │
│                                               │
│  Navigation State:                            │
│  - saveState = true for bottom nav            │
│  - restoreState = true for bottom nav         │
│  - launchSingleTop = true to avoid duplicates │
└───────────────────────────────────────────────┘
*/
```

---

## 15. Nested Navigation

### Nested Navigation Graphs

```kotlin
// ─── Nested Graph Definitions ──────────────────────────────

fun NavGraphBuilder.authNestedGraph(navController: NavHostController) {
    navigation(
        startDestination = Routes.LOGIN,
        route = "auth_graph"
    ) {
        composable(Routes.LOGIN) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate("main_graph") {
                        popUpTo("auth_graph") { inclusive = true }
                    }
                },
                onNavigateToRegister = {
                    navController.navigate(Routes.REGISTER)
                }
            )
        }

        composable(Routes.REGISTER) {
            RegisterScreen(
                onRegisterSuccess = {
                    navController.navigate("main_graph") {
                        popUpTo("auth_graph") { inclusive = true }
                    }
                },
                onNavigateBack = { navController.popBackStack() }
            )
        }
    }
}

fun NavGraphBuilder.mainNestedGraph(navController: NavHostController) {
    navigation(
        startDestination = Routes.DASHBOARD,
        route = "main_graph"
    ) {
        composable(Routes.DASHBOARD) {
            DashboardScreen(
                onNavigateToChat = { id ->
                    navController.navigate("chat_graph/$id")
                }
            )
        }

        // Nested chat graph within main
        navigation(
            startDestination = Routes.CHAT_LIST,
            route = "chat_graph/{conversationId}"
        ) {
            composable(
                route = Routes.CHAT_LIST,
                arguments = listOf(
                    navArgument("conversationId") {
                        type = NavType.StringType
                        nullable = true
                        defaultValue = null
                    }
                )
            ) { backStackEntry ->
                val conversationId = backStackEntry.arguments?.getString("conversationId")
                ChatListScreen(
                    initialConversationId = conversationId,
                    onChatClick = { id ->
                        navController.navigate("chat_detail_graph/$id")
                    }
                )
            }
        }

        // Nested chat detail graph
        navigation(
            startDestination = "chat_detail/{chatId}",
            route = "chat_detail_graph/{chatId}"
        ) {
            composable(
                route = "chat_detail/{chatId}",
                arguments = listOf(
                    navArgument("chatId") { type = NavType.StringType }
                )
            ) { backStackEntry ->
                val chatId = backStackEntry.arguments?.getString("chatId")
                ChatDetailScreen(
                    conversationId = chatId ?: "",
                    onNavigateBack = { navController.popBackStack() }
                )
            }
        }
    }
}

// ─── Main NavHost with Nested Graphs ───────────────────────

@Composable
fun NexusNavHost(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = "auth_graph"
    ) {
        authNestedGraph(navController)
        mainNestedGraph(navController)
    }
}

// ─── Scoped Navigation ─────────────────────────────────────

// Navigate within a nested graph
navController.navigate("main_graph/chat_list")

// Navigate up within nested graph
navController.navigateUp()

// Pop to parent graph
navController.popUpTo("main_graph")
```

### Nested Navigation Diagram

```
┌─────────────────────────────────────────────────────────┐
│                Nested Navigation Structure               │
│                                                         │
│  NavHost (startDestination = "auth_graph")             │
│  │                                                      │
│  ├── auth_graph                                         │
│  │   ├── login                                          │
│  │   └── register                                       │
│  │                                                      │
│  └── main_graph                                         │
│      ├── dashboard                                      │
│      ├── chat_graph/{conversationId}                    │
│      │   └── chat_list                                  │
│      │                                                  │
│      ├── chat_detail_graph/{chatId}                     │
│      │   └── chat_detail/{chatId}                       │
│      │                                                  │
│      ├── agents                                         │
│      │   └── agent_detail/{agentId}                     │
│      │                                                  │
│      ├── knowledge                                      │
│      │   └── document_detail/{documentId}               │
│      │                                                  │
│      └── settings                                       │
│          ├── settings/profile                           │
│          ├── settings/security                          │
│          └── settings/about                             │
└─────────────────────────────────────────────────────────┘
```

---

## 16. Modal Navigation

### Full-Screen Dialog Navigation

```kotlin
// ─── Modal Route Type ──────────────────────────────────────

sealed class ModalRoute(val route: String) {
    data object CreateConversation : ModalRoute("modal/create_conversation")
    data object EditProfile : ModalRoute("modal/edit_profile")
    data object SelectAgent : ModalRoute("modal/select_agent")
    data object ConfirmDelete : ModalRoute("modal/confirm_delete")
    data object ImagePreview : ModalRoute("modal/image_preview/{imageUrl}")
}

// ─── Dialog Composable ─────────────────────────────────────

@Composable
fun NexusNavHost(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD
    ) {
        // Regular screens
        composable(Routes.DASHBOARD) { /* ... */ }

        // Modal dialogs
        dialog(
            route = ModalRoute.CreateConversation.route,
            enterTransition = { fadeIn() + scaleIn(initialScale = 0.9f) },
            exitTransition = { fadeOut() + scaleOut(targetScale = 0.9f) },
            arguments = listOf(
                navArgument("imageUrl") {
                    type = NavType.StringType
                    nullable = true
                }
            )
        ) { backStackEntry ->
            CreateConversationDialog(
                onDismiss = { navController.popBackStack() },
                onConversationCreated = { id ->
                    navController.popBackStack()
                    navController.navigate(Routes.chatDetail(id))
                }
            )
        }

        dialog(
            route = ModalRoute.SelectAgent.route,
            enterTransition = { slideInVertically(initialOffsetY = { it }) + fadeIn() },
            exitTransition = { slideOutVertically(targetOffsetY = { it }) + fadeOut() }
        ) {
            SelectAgentDialog(
                onDismiss = { navController.popBackStack() },
                onAgentSelected = { agentId ->
                    navController.previousBackStackEntry
                        ?.savedStateHandle
                        ?.set("selected_agent_id", agentId)
                    navController.popBackStack()
                }
            )
        }

        dialog(
            route = ModalRoute.ConfirmDelete.route,
            enterTransition = { fadeIn() },
            exitTransition = { fadeOut() }
        ) { backStackEntry ->
            val itemName = backStackEntry.arguments?.getString("itemName") ?: "item"
            ConfirmDeleteDialog(
                itemName = itemName,
                onConfirm = {
                    navController.previousBackStackEntry
                        ?.savedStateHandle
                        ?.set("delete_confirmed", true)
                    navController.popBackStack()
                },
                onDismiss = { navController.popBackStack() }
            )
        }
    }
}

// ─── Dialog Composables ────────────────────────────────────

@Composable
fun CreateConversationDialog(
    onDismiss: () -> Unit,
    onConversationCreated: (String) -> Unit,
    viewModel: CreateConversationViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("New Conversation") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = state.title,
                    onValueChange = viewModel::updateTitle,
                    label = { Text("Title") },
                    singleLine = true
                )

                // Agent selection
                AgentDropdown(
                    selectedAgent = state.selectedAgent,
                    onAgentSelected = viewModel::selectAgent
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = { viewModel.createConversation() },
                enabled = state.isValid
            ) {
                Text("Create")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}

@Composable
fun SelectAgentDialog(
    onDismiss: () -> Unit,
    onAgentSelected: (String) -> Unit,
    viewModel: SelectAgentViewModel = hiltViewModel()
) {
    val agents by viewModel.agents.collectAsStateWithLifecycle()

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        dragHandle = { BottomSheetDefaults.DragHandle() }
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = "Select Agent",
                style = MaterialTheme.typography.titleLarge,
                modifier = Modifier.padding(bottom = 16.dp)
            )

            LazyColumn {
                items(agents) { agent ->
                    ListItem(
                        headlineContent = { Text(agent.name) },
                        supportingContent = { Text(agent.description ?: "") },
                        leadingContent = {
                            AsyncImage(
                                model = agent.avatarUrl,
                                contentDescription = null,
                                modifier = Modifier.size(40.dp).clip(CircleShape)
                            )
                        },
                        modifier = Modifier.clickable {
                            onAgentSelected(agent.id)
                        }
                    )
                }
            }
        }
    }
}
```

---

## 17. Sheet Navigation

### Bottom Sheet Navigation

```kotlin
// ─── Sheet Navigation Manager ──────────────────────────────

class SheetNavigator @Inject constructor() {
    private val _sheetState = MutableStateFlow<SheetState>(SheetState.Hidden)
    val sheetState: StateFlow<SheetState> = _sheetState.asStateFlow()

    fun showSheet(content: @Composable () -> Unit) {
        _sheetState.value = SheetState.Showing(content)
    }

    fun hideSheet() {
        _sheetState.value = SheetState.Hidden
    }

    sealed class SheetState {
        data object Hidden : SheetState()
        data class Showing(val content: @Composable () -> Unit) : SheetState()
    }
}

// ─── Sheet Host ────────────────────────────────────────────

@Composable
fun SheetHost(
    sheetNavigator: SheetNavigator,
    content: @Composable () -> Unit
) {
    val sheetState by sheetNavigator.sheetState.collectAsStateWithLifecycle()

    Box(modifier = Modifier.fillMaxSize()) {
        content()

        when (val state = sheetState) {
            is SheetNavigator.SheetState.Showing -> {
                ModalBottomSheet(
                    onDismissRequest = { sheetNavigator.hideSheet() }
                ) {
                    state.content()
                }
            }
            is SheetNavigator.SheetState.Hidden -> {
                // No sheet
            }
        }
    }
}

// ─── Usage ─────────────────────────────────────────────────

@Composable
fun ChatScreen(
    sheetNavigator: SheetNavigator = LocalSheetNavigator.current
) {
    // Trigger sheet from ViewModel event
    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is ChatEvent.ShowAgentSelector -> {
                    sheetNavigator.showSheet {
                        AgentSelectorSheet(
                            onAgentSelected = { id ->
                                viewModel.selectAgent(id)
                                sheetNavigator.hideSheet()
                            }
                        )
                    }
                }
                is ChatEvent.ShowAttachments -> {
                    sheetNavigator.showSheet {
                        AttachmentSheet(
                            onCameraClick = { /* Camera */ },
                            onGalleryClick = { /* Gallery */ },
                            onFileClick = { /* File picker */ }
                        )
                    }
                }
            }
        }
    }

    // Sheet content composables
    // ...
}
```

---

## 18. Navigation Animations

### Animation Configuration

```kotlin
// ─── Transition Specs ──────────────────────────────────────

object NavigationTransitions {
    val slideRight: EnterTransition
        get() = slideIntoContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Left,
            animationSpec = tween(300)
        )

    val slideRightExit: ExitTransition
        get() = slideOutOfContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Left,
            animationSpec = tween(300)
        )

    val slideLeft: EnterTransition
        get() = slideIntoContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Right,
            animationSpec = tween(300)
        )

    val slideLeftExit: ExitTransition
        get() = slideOutOfContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Right,
            animationSpec = tween(300)
        )

    val slideUp: EnterTransition
        get() = slideIntoContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Up,
            animationSpec = tween(400)
        )

    val slideDown: ExitTransition
        get() = slideOutOfContainer(
            towards = AnimatedContentTransitionScope.SlideDirection.Down,
            animationSpec = tween(400)
        )

    val fadeIn: EnterTransition
        get() = fadeIn(animationSpec = tween(200))

    val fadeOut: ExitTransition
        get() = fadeOut(animationSpec = tween(200))

    val scaleIn: EnterTransition
        get() = scaleIn(
            initialScale = 0.8f,
            animationSpec = tween(300)
        ) + fadeIn()

    val scaleOut: ExitTransition
        get() = scaleOut(
            targetScale = 0.8f,
            animationSpec = tween(300)
        ) + fadeOut()

    val sharedElement: SharedTransitionLayout
        get() = SharedTransitionLayout()
}

// ─── Animated NavHost ─────────────────────────────────────

@Composable
fun AnimatedNexusNavHost(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    SharedTransitionLayout {
        NavHost(
            navController = navController,
            startDestination = Routes.DASHBOARD,
            modifier = modifier,
            enterTransition = {
                fadeIn(animationSpec = tween(300)) +
                    slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Left)
            },
            exitTransition = {
                fadeOut(animationSpec = tween(300)) +
                    slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Left)
            },
            popEnterTransition = {
                fadeIn(animationSpec = tween(300)) +
                    slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Right)
            },
            popExitTransition = {
                fadeOut(animationSpec = tween(300)) +
                    slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Right)
            }
        ) {
            composable(
                route = Routes.DASHBOARD,
                enterTransition = { fadeIn(tween(200)) },
                exitTransition = { fadeOut(tween(200)) }
            ) {
                DashboardScreen()
            }

            composable(
                route = Routes.CHAT_DETAIL,
                enterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Left) },
                exitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Left) },
                popEnterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Right) },
                popExitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Right) }
            ) {
                ChatDetailScreen()
            }
        }
    }
}
```

### Animation Timing Table

```
┌─────────────────────────────────────────────────────────┐
│              Navigation Animation Timing                  │
│                                                         │
│  ┌──────────────────┬──────────┬─────────────────────┐  │
│  │ Transition       │ Duration │ Easing              │  │
│  ├──────────────────┼──────────┼─────────────────────┤  │
│  │ Slide Horizontal │ 300ms    │ EaseInOut           │  │
│  │ Slide Vertical   │ 400ms    │ EaseInOut           │  │
│  │ Fade             │ 200ms    │ Linear              │  │
│  │ Scale            │ 300ms    │ EaseOut             │  │
│  │ Shared Element   │ 350ms    │ Spring(damping=0.8) │  │
│  │ Bottom Sheet     │ 300ms    │ EaseOutCubic        │  │
│  │ Dialog           │ 250ms    │ EaseOut             │  │
│  └──────────────────┴──────────┴─────────────────────┘  │
│                                                         │
│  Performance Notes:                                     │
│  • Keep animations under 400ms for responsiveness       │
│  • Use hardware-accelerated properties only             │
│  • Test on low-end devices                             │
│  • Respect reduceMotion accessibility setting          │
└─────────────────────────────────────────────────────────┘
```

---

## 19. Navigation Testing

### Navigation Test Setup

```kotlin
@RunWith(AndroidJUnit4::class)
class NavigationTest {

    @get:Rule
    val composeRule = createComposeRule()

    private lateinit var navController: TestNavHostController

    @Before
    fun setup() {
        composeRule.setContent {
            navController = TestNavHostController(LocalContext.current)
            navController.graph = navController.createGraph(startDestination = "dashboard") {
                composable("dashboard") { /* Dashboard */ }
                composable("chat/{id}") { /* Chat */ }
                composable("settings") { /* Settings */ }
            }

            NexusNavHost(navController = navController)
        }
    }

    @Test
    fun startDestination_isDashboard() {
        assertEquals("dashboard", navController.currentDestination?.route)
    }

    @Test
    fun navigateToChat_setsCorrectRoute() {
        composeRule.onNodeWithText("Chat").performClick()
        assertEquals("chat/{id}", navController.currentDestination?.route)
    }

    @Test
    fun popBackStack_fromChat_returnsToDashboard() {
        // Navigate to chat
        navController.navigate("chat/123")

        // Pop back
        navController.popBackStack()

        assertEquals("dashboard", navController.currentDestination?.route)
    }

    @Test
    fun navigate_withSaveState_savesState() {
        // Navigate to settings
        navController.navigate("settings") {
            popUpTo("dashboard") { saveState = true }
            launchSingleTop = true
            restoreState = true
        }

        // Verify settings is current
        assertEquals("settings", navController.currentDestination?.route)
    }
}

// ─── Integration Test ──────────────────────────────────────

@RunWith(AndroidJUnit4::class)
class NavigationIntegrationTest {

    @get:Rule
    val composeRule = createAndroidComposeRule<MainActivity>()

    @Test
    fun appStarts_atDashboard() {
        composeRule.onNodeWithText("Dashboard").assertIsDisplayed()
    }

    @Test
    fun bottomNav_navigatesCorrectly() {
        // Click Chat tab
        composeRule.onNodeWithText("Chat").performClick()
        composeRule.waitForIdle()
        composeRule.onNodeWithText("Chat").assertIsSelected()

        // Click Agents tab
        composeRule.onNodeWithText("Agents").performClick()
        composeRule.waitForIdle()
        composeRule.onNodeWithText("Agents").assertIsSelected()
    }

    @Test
    fun backNavigation_worksCorrectly() {
        // Navigate to detail screen
        composeRule.onNodeWithText("New Chat").performClick()
        composeRule.waitForIdle()

        // Press back
        composeRule.onNodeWithContentDescription("Back").performClick()
        composeRule.waitForIdle()

        // Verify we're back at list
        composeRule.onNodeWithText("Dashboard").assertIsDisplayed()
    }
}
```

---

## 20. Navigation Accessibility

### Accessible Navigation

```kotlin
@Composable
fun AccessibleBottomBar(
    navController: NavHostController
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    NavigationBar {
        BottomNavItem.entries.forEach { item ->
            NavigationBarItem(
                selected = currentRoute == item.route,
                onClick = {
                    navController.navigate(item.route) {
                        popUpTo(navController.graph.findStartDestination().id) {
                            saveState = true
                        }
                        launchSingleTop = true
                        restoreState = true
                    }
                },
                icon = {
                    Icon(
                        imageVector = item.selectedIcon,
                        contentDescription = "${item.label} tab" +
                            if (currentRoute == item.route) ", selected" else ""
                    )
                },
                label = {
                    Text(
                        text = item.label,
                        modifier = Modifier.semantics {
                            stateDescription = if (currentRoute == item.route) {
                                "Selected"
                            } else {
                                "Not selected"
                            }
                        }
                    )
                }
            )
        }
    }
}

// ─── Accessible Back Button ────────────────────────────────

@Composable
fun AccessibleBackButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    IconButton(
        onClick = onClick,
        modifier = modifier.semantics {
            contentDescription = "Navigate back"
            role = Role.Button
        }
    ) {
        Icon(
            imageVector = Icons.Filled.ArrowBack,
            contentDescription = null // Decorative; parent has contentDescription
        )
    }
}

// ─── Accessible Navigation Announcements ───────────────────

@Composable
fun NavigationAnnouncer(
    currentRoute: String?
) {
    val announcement = when (currentRoute) {
        Routes.DASHBOARD -> "Dashboard screen"
        Routes.CHAT_LIST -> "Chat list screen"
        Routes.AGENTS -> "Agents screen"
        Routes.KNOWLEDGE -> "Knowledge screen"
        Routes.SETTINGS -> "Settings screen"
        else -> null
    }

    if (announcement != null) {
        // Announce screen change to TalkBack
        LaunchedEffect(currentRoute) {
            // Using semantics for screen reader
        }
    }
}

// ─── Navigation Accessibility Checklist ────────────────────

/*
┌──────────────────────────────────────────────────────┐
│         Navigation Accessibility Checklist            │
│                                                      │
│  ✓ All navigation elements have contentDescription   │
│  ✓ Selected state is announced                       │
│  ✓ Back button has descriptive label                 │
│  ✓ Screen changes are announced via liveRegion       │
│  ✓ Minimum touch target size (48x48dp)               │
│  ✓ Focus order is logical                            │
│  ✓ No navigation traps (always can go back)          │
│  ✓ Deep links work with keyboard navigation          │
│  ✓ Error states are announced                        │
│  ✓ Loading states have progress announcements        │
└──────────────────────────────────────────────────────┘
*/
```

---

## 21. Navigation Performance

### Lazy Screen Loading

```kotlin
// ─── Lazy Composition ──────────────────────────────────────

@Composable
fun LazyNavHost(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD,
        // Enable composition caching for faster navigation
        popEnterTransition = {
            fadeIn(animationSpec = tween(200))
        },
        popExitTransition = {
            fadeOut(animationSpec = tween(200))
        }
    ) {
        // Use remember for expensive screen creation
        composable(
            route = Routes.DASHBOARD,
            enterTransition = { fadeIn(tween(150)) },
            exitTransition = { fadeOut(tween(150)) }
        ) {
            key("dashboard") {
                DashboardScreen()
            }
        }

        composable(
            route = Routes.CHAT_DETAIL,
            enterTransition = { slideIntoContainer(AnimatedContentTransitionScope.SlideDirection.Left, tween(200)) },
            exitTransition = { slideOutOfContainer(AnimatedContentTransitionScope.SlideDirection.Left, tween(200)) }
        ) { backStackEntry ->
            key(backStackEntry.arguments?.getString("conversationId")) {
                ChatDetailScreen(
                    conversationId = backStackEntry.arguments?.getString("conversationId") ?: ""
                )
            }
        }
    }
}

// ─── Prefetching Data ──────────────────────────────────────

@HiltViewModel
class NavigationPerfViewModel @Inject constructor(
    private val repository: ConversationRepository
) : ViewModel() {

    init {
        // Prefetch data for likely next screens
        viewModelScope.launch {
            repository.prefetchConversations()
        }
    }
}

// ─── Navigation Performance Metrics ────────────────────────

class NavigationPerfTracker @Inject constructor() {
    private var lastNavigationTime = 0L

    fun onNavigationStart() {
        lastNavigationTime = System.nanoTime()
    }

    fun onNavigationComplete() {
        val duration = System.nanoTime() - lastNavigationTime
        val durationMs = duration / 1_000_000

        if (durationMs > 500) {
            Timber.w("Slow navigation: ${durationMs}ms")
        }
    }
}
```

### Performance Benchmarks

| Metric | Target | Warning | Critical |
|--------|--------|---------|----------|
| Screen transition | < 200ms | > 300ms | > 500ms |
| First composition | < 500ms | > 800ms | > 1200ms |
| Recomposition | < 16ms | > 32ms | > 64ms |
| Deep link handling | < 300ms | > 500ms | > 800ms |
| Back stack restore | < 100ms | > 200ms | > 400ms |

---

## 22. Navigation Error Handling

### Unknown Route Handling

```kotlin
@Composable
fun NexusNavHost(navController: NavHostController) {
    NavHost(
        navController = navController,
        startDestination = Routes.DASHBOARD
    ) {
        // All known routes...

        // Catch unknown routes
        composable(
            route = "{anyRoute}",
            arguments = listOf(
                navArgument("anyRoute") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val unknownRoute = backStackEntry.arguments?.getString("anyRoute")

            // Log error
            LaunchedEffect(unknownRoute) {
                Timber.e("Unknown route: $unknownRoute")
            }

            NotFoundScreen(
                route = unknownRoute ?: "",
                onNavigateHome = {
                    navController.navigate(Routes.DASHBOARD) {
                        popUpTo(0) { inclusive = true }
                    }
                },
                onNavigateBack = { navController.popBackStack() }
            )
        }
    }
}

@Composable
fun NotFoundScreen(
    route: String,
    onNavigateHome: () -> Unit,
    onNavigateBack: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            Icons.Filled.ErrorOutline,
            contentDescription = null,
            modifier = Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.error
        )

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            text = "Page Not Found",
            style = MaterialTheme.typography.headlineMedium
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "The route '$route' doesn't exist.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(24.dp))

        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            OutlinedButton(onClick = onNavigateBack) {
                Text("Go Back")
            }

            Button(onClick = onNavigateHome) {
                Text("Go Home")
            }
        }
    }
}
```

---

## 23. Navigation from Notifications

### Notification Deep Link Handler

```kotlin
class NotificationNavigator @Inject constructor(
    private val deepLinkHandler: DeepLinkHandler
) {
    fun handleNotificationIntent(
        intent: Intent,
        navController: NavHostController
    ) {
        val data = intent.data ?: return
        val deepLinkResult = deepLinkHandler.handleIntent(data)

        when (deepLinkResult) {
            is DeepLinkResult.Valid -> {
                navController.navigate(deepLinkResult.route) {
                    popUpTo(navController.graph.findStartDestination().id) {
                        saveState = true
                    }
                    launchSingleTop = true
                    restoreState = true
                }
            }
            is DeepLinkResult.Invalid -> {
                // Fallback to main screen
                navController.navigate(Routes.DASHBOARD) {
                    popUpTo(0) { inclusive = true }
                }
            }
            is DeepLinkResult.Expired -> {
                // Show expired link message
            }
        }
    }
}

// ─── Notification Builder with Deep Link ───────────────────

class NotificationHelper @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun createMessageNotification(
        conversationId: String,
        title: String,
        message: String
    ) {
        val intent = Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("https://nexus-ai.com/app/chat/$conversationId")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }

        val pendingIntent = PendingIntent.getActivity(
            context,
            conversationId.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val notification = NotificationCompat.Builder(context, "messages")
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(title)
            .setContentText(message)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)
            .build()

        val manager = context.getSystemService<NotificationManager>()
        manager?.notify(conversationId.hashCode(), notification)
    }
}
```

---

## 24. Navigation from Share Intent

### Share Intent Handler

```kotlin
class ShareIntentHandler @Inject constructor() {

    data class ShareData(
        val text: String?,
        val images: List<Uri>?,
        val files: List<Uri>?
    )

    fun handleShareIntent(
        intent: Intent,
        navController: NavHostController
    ) {
        when (intent.action) {
            Intent.ACTION_SEND -> handleSingleShare(intent, navController)
            Intent.ACTION_SEND_MULTIPLE -> handleMultipleShare(intent, navController)
        }
    }

    private fun handleSingleShare(intent: Intent, navController: NavHostController) {
        val sharedText = intent.getStringExtra(Intent.EXTRA_TEXT)
        val sharedImage = intent.getParcelableExtra<Uri>(Intent.EXTRA_STREAM)

        val shareData = ShareData(
            text = sharedText,
            images = sharedImage?.let { listOf(it) },
            files = null
        )

        navigateToShareScreen(navController, shareData)
    }

    private fun handleMultipleShare(intent: Intent, navController: NavHostController) {
        val sharedText = intent.getStringExtra(Intent.EXTRA_TEXT)
        val sharedImages = intent.getParcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)

        val shareData = ShareData(
            text = sharedText,
            images = sharedImages,
            files = null
        )

        navigateToShareScreen(navController, shareData)
    }

    private fun navigateToShareScreen(
        navController: NavHostController,
        shareData: ShareData
    ) {
        // Store share data and navigate
        navController.currentBackStackEntry
            ?.savedStateHandle
            ?.apply {
                set("share_text", shareData.text)
                set("share_images", shareData.images)
            }

        navController.navigate(Routes.CREATE_CONVERSATION) {
            popUpTo(Routes.DASHBOARD)
            launchSingleTop = true
        }
    }
}

// ─── AndroidManifest Share Intent Filter ───────────────────

/*
<activity
    android:name=".MainActivity"
    android:launchMode="singleTop">

    <intent-filter>
        <action android:name="android.intent.action.SEND" />
        <category android:name="android.intent.category.DEFAULT" />
        <data android:mimeType="text/plain" />
    </intent-filter>

    <intent-filter>
        <action android:name="android.intent.action.SEND" />
        <category android:name="android.intent.category.DEFAULT" />
        <data android:mimeType="image/*" />
    </intent-filter>

    <intent-filter>
        <action android:name="android.intent.action.SEND_MULTIPLE" />
        <category android:name="android.intent.category.DEFAULT" />
        <data android:mimeType="image/*" />
    </intent-filter>
</activity>
*/
```

---

## 25. Navigation from Widget

### Widget Navigation

```kotlin
class WidgetNavigationHelper @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun createChatPendingIntent(conversationId: String): PendingIntent {
        val intent = Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("nexusai://open/chat/$conversationId")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }

        return PendingIntent.getActivity(
            context,
            conversationId.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    fun createDashboardPendingIntent(): PendingIntent {
        val intent = Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("nexusai://open/dashboard")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }

        return PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    fun createNewChatPendingIntent(): PendingIntent {
        val intent = Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("nexusai://open/create_chat")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }

        return PendingIntent.getActivity(
            context,
            -1,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }
}

// ─── App Widget Implementation ─────────────────────────────

class NexusWidget : AppWidgetProvider() {

    @Inject
    lateinit var widgetNavigationHelper: WidgetNavigationHelper

    override fun onUpdate(
        context: Context,
        appWidgetManager: AppWidgetManager,
        appWidgetIds: IntArray
    ) {
        for (appWidgetId in appWidgetIds) {
            updateAppWidget(context, appWidgetManager, appWidgetId)
        }
    }

    private fun updateAppWidget(
        context: Context,
        appWidgetManager: AppWidgetManager,
        appWidgetId: Int
    ) {
        val views = RemoteViews(context.packageName, R.layout.widget_layout)

        // New Chat button
        views.setOnClickPendingIntent(
            R.id.widget_new_chat_button,
            widgetNavigationHelper.createNewChatPendingIntent()
        )

        // Dashboard button
        views.setOnClickPendingIntent(
            R.id.widget_dashboard_button,
            widgetNavigationHelper.createDashboardPendingIntent()
        )

        // Recent chat items (using RemoteViewsAdapter)
        val intent = Intent(context, WidgetRemoteViewsService::class.java)
        views.setRemoteAdapter(R.id.widget_list_view, intent)

        // Handle item clicks
        val clickIntent = Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("nexusai://open/chat/")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK
        }
        val clickPendingIntent = PendingIntent.getActivity(
            context, 0, clickIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_MUTABLE
        )
        views.setPendingIntentTemplate(R.id.widget_list_view, clickPendingIntent)

        appWidgetManager.updateAppWidget(appWidgetId, views)
    }
}

// ─── Widget Remote Views Service ───────────────────────────

class WidgetRemoteViewsService : RemoteViewsService() {
    override fun onGetViewFactory(intent: Intent): RemoteViewsFactory {
        return WidgetRemoteViewsFactory(applicationContext)
    }
}

class WidgetRemoteViewsFactory(
    private val context: Context
) : RemoteViewsService.RemoteViewsFactory {

    private var conversations: List<Conversation> = emptyList()

    override fun onCreate() {
        // Load recent conversations
        CoroutineScope(Dispatchers.IO).launch {
            // Load from repository
        }
    }

    override fun onDataSetChanged() {
        // Refresh data
    }

    override fun getCount(): Int = conversations.size

    override fun getViewAt(position: Int): RemoteViews {
        val conversation = conversations[position]

        return RemoteViews(context.packageName, R.layout.widget_chat_item).apply {
            setTextViewText(R.id.widget_chat_title, conversation.title)
            setTextViewText(R.id.widget_chat_preview, conversation.lastMessage)
            setTextViewText(R.id.widget_chat_time, formatTime(conversation.lastMessageTime))

            // Fill in click intent template data
            val fillInIntent = Intent().apply {
                putExtra("conversation_id", conversation.id)
            }
            setOnClickFillInIntent(R.id.widget_chat_item, fillInIntent)
        }
    }

    override fun getLoadingView(): RemoteViews? = null
    override fun getViewTypeCount(): Int = 1
    override fun getItemId(position: Int): Long = conversations[position].id.hashCode().toLong()
    override fun hasStableIds(): Boolean = true

    private fun formatTime(timestamp: Long): String {
        val now = System.currentTimeMillis()
        val diff = now - timestamp

        return when {
            diff < 60_000 -> "now"
            diff < 3_600_000 -> "${diff / 60_000}m"
            diff < 86_400_000 -> "${diff / 3_600_000}h"
            else -> "${diff / 86_400_000}d"
        }
    }
}
```

---

## Summary

| Feature | Implementation | Library |
|---------|---------------|---------|
| **Navigation Graph** | NavHost + composable routes | Navigation Compose |
| **Type-Safe Routes** | Sealed class / Kotlin serialization | Navigation Compose |
| **Bottom Navigation** | NavigationBar + NavigationBarItem | Material 3 |
| **Top App Bar** | TopAppBar / LargeTopAppBar | Material 3 |
| **Navigation Drawer** | ModalNavigationDrawer | Material 3 |
| **Deep Links** | intent-filter + navDeepLink | Navigation + Manifest |
| **Sheet Navigation** | ModalBottomSheet | Material 3 |
| **Animations** | enterTransition/exitTransition | Navigation Compose |
| **Nested Navigation** | navigation() builder | Navigation Compose |
| **State Preservation** | saveState/restoreState | Navigation Compose |
| **Back Stack** | popBackStack/navigateUp | Navigation Compose |
| **Testing** | TestNavHostController | Navigation Testing |
| **Accessibility** | contentDescription + semantics | Compose |
| **Performance** | Animation duration + caching | Custom |
