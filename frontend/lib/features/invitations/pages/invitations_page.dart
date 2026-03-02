import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/invitations_cubit.dart';

class InvitationsPage extends StatelessWidget {
  const InvitationsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<String?>(
      future: sl<ApiClient>().getRole(),
      builder: (context, snapshot) {
        final role = snapshot.data ?? '';
        return BlocProvider(
          create: (_) {
            final cubit = sl<InvitationsCubit>();
            if (role == 'agency') {
              cubit.loadMyOutgoing();
            } else {
              cubit.loadMyInvitations();
            }
            return cubit;
          },
          child: _InvitationsView(role: role),
        );
      },
    );
  }
}

class _InvitationsView extends StatelessWidget {
  final String role;
  const _InvitationsView({required this.role});

  @override
  Widget build(BuildContext context) {
    final isAgency = role == 'agency';
    final title = isAgency ? 'Sent Invitations' : 'Invitations';

    return Scaffold(
      appBar: AppBar(title: Text(title)),
      drawer: const AppDrawer(),
      body: BlocConsumer<InvitationsCubit, InvitationsState>(
        listener: (context, state) {
          if (state is InvitationActionSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
          }
        },
        builder: (context, state) {
          if (state is InvitationsLoading || state is InvitationsInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is InvitationsError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(state.message,
                      style: const TextStyle(color: Colors.red)),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () {
                      if (isAgency) {
                        context.read<InvitationsCubit>().loadMyOutgoing();
                      } else {
                        context.read<InvitationsCubit>().loadMyInvitations();
                      }
                    },
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          if (state is InvitationsLoaded) {
            if (state.invitations.isEmpty) {
              return Center(
                child: Text(
                  isAgency ? 'No sent invitations.' : 'No invitations.',
                  style: const TextStyle(color: Colors.grey),
                ),
              );
            }
            return ListView.builder(
              itemCount: state.invitations.length,
              itemBuilder: (ctx, i) {
                final inv = state.invitations[i] as Map<String, dynamic>;
                return _InvitationCard(
                  invitation: inv,
                  isReceived: !isAgency,
                  onRespond: !isAgency && inv['status'] == 'pending'
                      ? (status) => context
                          .read<InvitationsCubit>()
                          .respondToInvitation(inv['id'] as String, status)
                      : null,
                );
              },
            );
          }
          return const SizedBox();
        },
      ),
    );
  }
}

class _InvitationCard extends StatelessWidget {
  final Map<String, dynamic> invitation;
  final bool isReceived;
  final void Function(String status)? onRespond;

  const _InvitationCard({
    required this.invitation,
    required this.isReceived,
    this.onRespond,
  });

  @override
  Widget build(BuildContext context) {
    final status = invitation['status'] as String? ?? 'pending';
    final message = invitation['message'] as String? ?? '';

    String displayName;
    if (isReceived) {
      final agency = invitation['agency'] as Map<String, dynamic>? ?? {};
      displayName = agency['company_name'] as String? ?? 'Unknown Agency';
    } else {
      final model = invitation['model'] as Map<String, dynamic>? ?? {};
      displayName =
          '${model['first_name'] ?? ''} ${model['last_name'] ?? ''}'.trim();
      if (displayName.isEmpty) displayName = 'Unknown Model';
    }

    Color statusColor;
    switch (status) {
      case 'accepted':
        statusColor = Colors.green;
        break;
      case 'rejected':
        statusColor = Colors.red;
        break;
      default:
        statusColor = Colors.orange;
    }

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    isReceived
                        ? 'From: $displayName'
                        : 'To: $displayName',
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: statusColor.withAlpha(30),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: statusColor),
                  ),
                  child: Text(
                    status.toUpperCase(),
                    style: TextStyle(color: statusColor, fontSize: 11),
                  ),
                ),
              ],
            ),
            if (message.isNotEmpty) ...[
              const SizedBox(height: 6),
              Text(message, style: const TextStyle(color: Colors.grey)),
            ],
            if (isReceived && status == 'pending' && onRespond != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  OutlinedButton(
                    onPressed: () => onRespond!('rejected'),
                    style:
                        OutlinedButton.styleFrom(foregroundColor: Colors.red),
                    child: const Text('Decline'),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: () => onRespond!('accepted'),
                    child: const Text('Accept'),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}
