import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../app/router_paths.dart';
import '../../../../app/session.dart';
import '../../../../core/extensions/theme_context_x.dart';
import '../../../../core/widgets/design_system/app_button.dart';
import '../../../../core/widgets/design_system/app_card.dart';
import '../../domain/entities/auth_action_entity.dart';
import '../providers/auth_providers.dart';

enum AuthMode {
  login,
  register,
}

class AuthPage extends ConsumerWidget {
  const AuthPage({
    required this.mode,
    super.key,
    this.initialUsername,
    this.registrationSuccess = false,
  });

  final AuthMode mode;
  final String? initialUsername;
  final bool registrationSuccess;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: _AuthBackground(
        child: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final isWide = constraints.maxWidth >= 1040;

              return Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 1240),
                  child: Padding(
                    padding: EdgeInsets.symmetric(
                      horizontal: constraints.maxWidth >= 720 ? 32 : 20,
                      vertical: 20,
                    ),
                    child: isWide
                        ? Row(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              Expanded(
                                flex: 5,
                                child: _AuthHeroPanel(mode: mode),
                              ),
                              const SizedBox(width: 28),
                              Expanded(
                                flex: 4,
                                child: _AuthFormColumn(
                                  mode: mode,
                                  initialUsername: initialUsername,
                                  registrationSuccess: registrationSuccess,
                                ),
                              ),
                            ],
                          )
                        : SingleChildScrollView(
                            child: Column(
                              children: [
                                _AuthHeroPanel(mode: mode, compact: true),
                                const SizedBox(height: 20),
                                _AuthFormColumn(
                                  mode: mode,
                                  initialUsername: initialUsername,
                                  registrationSuccess: registrationSuccess,
                                ),
                              ],
                            ),
                          ),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}

class _AuthBackground extends StatelessWidget {
  const _AuthBackground({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;

    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: <Color>[
            scheme.surface,
            scheme.surface.withOpacity(0.98),
            const Color(0xFFF1F7FF),
          ],
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            top: -80,
            left: -50,
            child: _GlowBlob(color: scheme.primary.withOpacity(0.16), size: 220),
          ),
          Positioned(
            top: 120,
            right: -40,
            child: _GlowBlob(color: const Color(0xFF8DD6FF).withOpacity(0.18), size: 180),
          ),
          Positioned(
            bottom: -70,
            right: 120,
            child: _GlowBlob(color: const Color(0xFF0C63E7).withOpacity(0.08), size: 260),
          ),
          child,
        ],
      ),
    );
  }
}

class _GlowBlob extends StatelessWidget {
  const _GlowBlob({required this.color, required this.size});

  final Color color;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: color,
      ),
    );
  }
}

class _AuthHeroPanel extends StatelessWidget {
  const _AuthHeroPanel({
    required this.mode,
    this.compact = false,
  });

  final AuthMode mode;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final tokens = context.tokens;
    final headline = mode == AuthMode.login
        ? 'Sign in with a passkey-first flow.'
        : 'Create a secure account without passwords.';
    final description = mode == AuthMode.login
        ? 'CardPhanton keeps authentication browser-native, fast, and tied to your device or security key.'
        : 'Register a username, pair your first passkey, and keep the whole journey cookie-backed and simple.';

    return AppCard(
      padding: EdgeInsets.zero,
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(tokens.radius.lg),
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: <Color>[
              scheme.primary.withOpacity(0.95),
              const Color(0xFF083C7C),
            ],
          ),
        ),
        padding: EdgeInsets.all(compact ? 24 : 36),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.12),
                    borderRadius: BorderRadius.circular(tokens.radius.pill),
                    border: Border.all(color: Colors.white.withOpacity(0.16)),
                  ),
                  child: const Text(
                    'CardPhanton secure access',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.w600,
                      letterSpacing: 0.2,
                    ),
                  ),
                ),
                SizedBox(height: compact ? 18 : 26),
                Text(
                  headline,
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        color: Colors.white,
                        height: 1.08,
                      ),
                ),
                SizedBox(height: compact ? 12 : 16),
                Text(
                  description,
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                        color: Colors.white.withOpacity(0.86),
                        height: 1.5,
                      ),
                ),
              ],
            ),
            SizedBox(height: compact ? 24 : 36),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: const [
                _ValueChip(label: 'Passkey-first'),
                _ValueChip(label: 'Browser-native'),
                _ValueChip(label: 'Cookie-backed'),
              ],
            ),
            SizedBox(height: compact ? 24 : 36),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _FeatureLine(
                  title: 'One-tap browser ceremony',
                  subtitle: "Uses the Go server's registration and authentication endpoints directly.",
                ),
                const SizedBox(height: 14),
                _FeatureLine(
                  title: 'Fast route transitions',
                  subtitle: 'Successful login immediately updates the app session and redirects to the dashboard.',
                ),
                const SizedBox(height: 14),
                _FeatureLine(
                  title: 'Designed for the app shell',
                  subtitle: 'The layout stays aligned with the existing tokens, cards, and motion system.',
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _FeatureLine extends StatelessWidget {
  const _FeatureLine({
    required this.title,
    required this.subtitle,
  });

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(top: 5),
          width: 10,
          height: 10,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: Colors.white,
                    ),
              ),
              const SizedBox(height: 4),
              Text(
                subtitle,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.white.withOpacity(0.76),
                      height: 1.45,
                    ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _ValueChip extends StatelessWidget {
  const _ValueChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.10),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: Colors.white.withOpacity(0.14)),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: Colors.white,
            ),
      ),
    );
  }
}

class _AuthFormColumn extends StatelessWidget {
  const _AuthFormColumn({
    required this.mode,
    required this.initialUsername,
    required this.registrationSuccess,
  });

  final AuthMode mode;
  final String? initialUsername;
  final bool registrationSuccess;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(28),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _AuthHeader(mode: mode),
          if (registrationSuccess)
            Padding(
              padding: const EdgeInsets.only(top: 20),
              child: _InfoBanner(
                color: const Color(0xFFE8FFF3),
                borderColor: const Color(0xFFB1F0CB),
                foregroundColor: const Color(0xFF0A7A3B),
                icon: Icons.verified_rounded,
                message: 'Registration completed. Your passkey is ready for sign in.',
              ),
            ),
          const SizedBox(height: 24),
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 240),
            switchInCurve: Curves.easeOutCubic,
            switchOutCurve: Curves.easeInCubic,
            child: mode == AuthMode.login
                ? _LoginForm(initialUsername: initialUsername)
                : _RegisterForm(initialUsername: initialUsername),
          ),
          const SizedBox(height: 20),
          _AuthRouteSwitcher(mode: mode),
        ],
      ),
    );
  }
}

class _AuthHeader extends StatelessWidget {
  const _AuthHeader({required this.mode});

  final AuthMode mode;

  @override
  Widget build(BuildContext context) {
    final title = mode == AuthMode.login ? 'Welcome back' : 'Create your account';
    final subtitle = mode == AuthMode.login
        ? 'Use a passkey or security key to sign in.'
        : 'Register a username and bind the first passkey to this workspace.';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
        ),
        const SizedBox(height: 8),
        Text(
          subtitle,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.68),
                height: 1.45,
              ),
        ),
      ],
    );
  }
}

class _LoginForm extends ConsumerStatefulWidget {
  const _LoginForm({this.initialUsername});

  final String? initialUsername;

  @override
  ConsumerState<_LoginForm> createState() => _LoginFormState();
}

class _LoginFormState extends ConsumerState<_LoginForm> {
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  late final TextEditingController _usernameController;
  String? _errorMessage;
  bool _isLoading = false;
  bool _attemptedConditional = false;

  @override
  void initState() {
    super.initState();
    _usernameController = TextEditingController(text: widget.initialUsername ?? '');
    _attemptConditionalLogin();
  }

  @override
  void dispose() {
    _usernameController.dispose();
    super.dispose();
  }

  Future<void> _attemptConditionalLogin() async {
    if (_attemptedConditional || _usernameController.text.trim().isEmpty) {
      return;
    }
    _attemptedConditional = true;

    await Future<void>.delayed(const Duration(milliseconds: 250));
    if (!mounted) {
      return;
    }

    try {
      await _runAuthentication(conditional: true);
    } catch (_) {
      // Ignore autofill failures and leave the form available for manual sign-in.
    }
  }

  Future<void> _runAuthentication({required bool conditional}) async {
    final username = _usernameController.text.trim();
    if (username.isEmpty) {
      setState(() => _errorMessage = 'Username is required.');
      return;
    }

    setState(() {
      _errorMessage = null;
      _isLoading = true;
    });

    try {
      final authUseCase = ref.read(authUseCaseProvider);
      final browserClient = ref.read(webAuthnBrowserClientProvider);

      final options = conditional
          ? await authUseCase.generateConditionalOptions()
          : await authUseCase.generateAuthenticationOptions(username);

      final credential = await browserClient.getAssertion(
        options,
        conditional: conditional,
      );

      final result = await authUseCase.verifyAuthentication(
        username: username,
        credential: credential,
      );

      if (!result.success) {
        throw StateError(result.message.isEmpty ? 'Login failed.' : result.message);
      }

      ref.read(authSessionProvider).updateAuth(
            isAuthenticated: true,
            role: UserRole.user,
            username: result.username ?? username,
          );

      if (mounted) {
        context.go(RouterPaths.dashboard);
      }
    } catch (error) {
      if (mounted) {
        setState(() => _errorMessage = _messageFromError(error));
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  String _messageFromError(Object error) {
    final message = error.toString();
    if (message.startsWith('StateError: ')) {
      return message.replaceFirst('StateError: ', '');
    }
    return message;
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _AuthField(
            controller: _usernameController,
            label: 'Username',
            hintText: 'johndoe',
            prefixIcon: Icons.badge_outlined,
            textInputAction: TextInputAction.done,
            autofillHints: const <String>[AutofillHints.username],
            validator: (value) {
              final cleaned = value?.trim() ?? '';
              if (cleaned.isEmpty) {
                return 'Enter your username.';
              }
              if (cleaned.length < 3) {
                return 'Username must be at least 3 characters.';
              }
              return null;
            },
            onSubmitted: (_) {
              if (_formKey.currentState?.validate() ?? false) {
                unawaited(_runAuthentication(conditional: false));
              }
            },
          ),
          const SizedBox(height: 16),
          _InfoBanner(
            color: const Color(0xFFF5F9FF),
            borderColor: const Color(0xFFDCEAFF),
            foregroundColor: const Color(0xFF0C4DA7),
            icon: Icons.lock_outline_rounded,
            message: 'Authentication uses the browser passkey prompt and sets a secure cookie on success.',
          ),
          if (_errorMessage != null) ...[
            const SizedBox(height: 16),
            _InfoBanner(
              color: const Color(0xFFFFF1F1),
              borderColor: const Color(0xFFF4C8C8),
              foregroundColor: const Color(0xFFB42318),
              icon: Icons.error_outline_rounded,
              message: _errorMessage!,
            ),
          ],
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            height: 50,
            child: AppButton(
              label: _isLoading ? 'Verifying passkey...' : 'Continue with passkey',
              isLoading: _isLoading,
              onPressed: () {
                if (_formKey.currentState?.validate() ?? false) {
                  unawaited(_runAuthentication(conditional: false));
                }
              },
            ),
          ),
          const SizedBox(height: 14),
          TextButton.icon(
            onPressed: _isLoading
                ? null
                : () {
                    if (_formKey.currentState?.validate() ?? false) {
                      unawaited(_runAuthentication(conditional: true));
                    }
                  },
            icon: const Icon(Icons.bolt_rounded, size: 18),
            label: const Text('Try autofill passkey'),
          ),
        ],
      ),
    );
  }
}

class _RegisterForm extends ConsumerStatefulWidget {
  const _RegisterForm({this.initialUsername});

  final String? initialUsername;

  @override
  ConsumerState<_RegisterForm> createState() => _RegisterFormState();
}

class _RegisterFormState extends ConsumerState<_RegisterForm> {
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  late final TextEditingController _usernameController;
  late final TextEditingController _nicknameController;
  String? _errorMessage;
  String? _successMessage;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _usernameController = TextEditingController(text: widget.initialUsername ?? '');
    _nicknameController = TextEditingController(text: 'This device');
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _nicknameController.dispose();
    super.dispose();
  }

  Future<void> _runRegistration() async {
    final username = _usernameController.text.trim();
    final nickname = _nicknameController.text.trim();

    if (username.isEmpty) {
      setState(() => _errorMessage = 'Username is required.');
      return;
    }

    setState(() {
      _errorMessage = null;
      _successMessage = null;
      _isLoading = true;
    });

    try {
      final authUseCase = ref.read(authUseCaseProvider);
      final browserClient = ref.read(webAuthnBrowserClientProvider);

      final options = await authUseCase.generateRegistrationOptions(username);
      final credential = await browserClient.createCredential(options);

      final result = await authUseCase.verifyRegistration(
        username: username,
        credential: credential,
        deviceNickname: nickname.isEmpty ? null : nickname,
      );

      if (!result.success) {
        throw StateError(result.message.isEmpty ? 'Registration failed.' : result.message);
      }

      if (mounted) {
        setState(() {
          _successMessage = result.message.isEmpty
              ? 'Registration completed. You can now sign in.'
              : result.message;
        });

        unawaited(Future<void>.delayed(const Duration(milliseconds: 450), () {
          if (!mounted) {
            return;
          }
          context.go(
            '${RouterPaths.login}?username=${Uri.encodeComponent(username)}&registered=1',
          );
        }));
      }
    } catch (error) {
      if (mounted) {
        setState(() => _errorMessage = _messageFromError(error));
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  String _messageFromError(Object error) {
    final message = error.toString();
    if (message.startsWith('StateError: ')) {
      return message.replaceFirst('StateError: ', '');
    }
    return message;
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _AuthField(
            controller: _usernameController,
            label: 'Username',
            hintText: 'johndoe',
            prefixIcon: Icons.badge_outlined,
            textInputAction: TextInputAction.next,
            autofillHints: const <String>[AutofillHints.username],
            validator: (value) {
              final cleaned = value?.trim() ?? '';
              if (cleaned.isEmpty) {
                return 'Enter a username.';
              }
              if (cleaned.length < 3) {
                return 'Username must be at least 3 characters.';
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          _AuthField(
            controller: _nicknameController,
            label: 'Device nickname',
            hintText: 'MacBook Pro',
            prefixIcon: Icons.devices_outlined,
            textInputAction: TextInputAction.done,
            autofillHints: const <String>[AutofillHints.nickname],
            validator: (value) {
              final cleaned = value?.trim() ?? '';
              if (cleaned.isEmpty) {
                return 'Give this passkey a friendly device name.';
              }
              return null;
            },
            onSubmitted: (_) {
              if (_formKey.currentState?.validate() ?? false) {
                unawaited(_runRegistration());
              }
            },
          ),
          const SizedBox(height: 16),
          _InfoBanner(
            color: const Color(0xFFF5F9FF),
            borderColor: const Color(0xFFDCEAFF),
            foregroundColor: const Color(0xFF0C4DA7),
            icon: Icons.verified_user_rounded,
            message: 'The Go server creates the publicKey options and the browser returns the attestation for verification.',
          ),
          if (_errorMessage != null) ...[
            const SizedBox(height: 16),
            _InfoBanner(
              color: const Color(0xFFFFF1F1),
              borderColor: const Color(0xFFF4C8C8),
              foregroundColor: const Color(0xFFB42318),
              icon: Icons.error_outline_rounded,
              message: _errorMessage!,
            ),
          ],
          if (_successMessage != null) ...[
            const SizedBox(height: 16),
            _InfoBanner(
              color: const Color(0xFFE8FFF3),
              borderColor: const Color(0xFFB1F0CB),
              foregroundColor: const Color(0xFF0A7A3B),
              icon: Icons.check_circle_outline_rounded,
              message: _successMessage!,
            ),
          ],
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            height: 50,
            child: AppButton(
              label: _isLoading ? 'Creating passkey...' : 'Register passkey',
              isLoading: _isLoading,
              onPressed: () {
                if (_formKey.currentState?.validate() ?? false) {
                  unawaited(_runRegistration());
                }
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _AuthField extends StatelessWidget {
  const _AuthField({
    required this.controller,
    required this.label,
    required this.hintText,
    required this.prefixIcon,
    required this.textInputAction,
    required this.validator,
    required this.autofillHints,
    this.onSubmitted,
  });

  final TextEditingController controller;
  final String label;
  final String hintText;
  final IconData prefixIcon;
  final TextInputAction textInputAction;
  final String? Function(String?) validator;
  final List<String> autofillHints;
  final ValueChanged<String>? onSubmitted;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;

    return TextFormField(
      controller: controller,
      validator: validator,
      autofillHints: autofillHints,
      textInputAction: textInputAction,
      onFieldSubmitted: onSubmitted,
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        prefixIcon: Icon(prefixIcon),
        filled: true,
        fillColor: scheme.surface.withOpacity(0.72),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.outlineVariant.withOpacity(0.7)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.outlineVariant.withOpacity(0.7)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.primary, width: 1.6),
        ),
      ),
    );
  }
}

class _InfoBanner extends StatelessWidget {
  const _InfoBanner({
    required this.color,
    required this.borderColor,
    required this.foregroundColor,
    required this.icon,
    required this.message,
  });

  final Color color;
  final Color borderColor;
  final Color foregroundColor;
  final IconData icon;
  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: borderColor),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 18, color: foregroundColor),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: foregroundColor,
                    height: 1.45,
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AuthRouteSwitcher extends StatelessWidget {
  const _AuthRouteSwitcher({required this.mode});

  final AuthMode mode;

  @override
  Widget build(BuildContext context) {
    final isLogin = mode == AuthMode.login;
    final prompt = isLogin ? "Don't have an account?" : 'Already have an account?';
    final action = isLogin ? 'Create one' : 'Sign in';

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(
          prompt,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.68),
              ),
        ),
        TextButton(
          onPressed: () {
            context.go(isLogin ? RouterPaths.register : RouterPaths.login);
          },
          child: Text(action),
        ),
      ],
    );
  }
}
