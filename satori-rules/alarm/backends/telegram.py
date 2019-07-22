# -*- coding: utf-8 -*-


# -- stdlib --
# -- third party --
from telegram.ext import CommandHandler, CallbackQueryHandler, Updater
import telegram

# -- own --
from backend import Backend


# -- code --
class TelegramBackend(Backend):
    def __init__(self, conf):
        super(TelegramBackend, self).__init__(conf)
        self.init_telegram_bot()

    def init_telegram_bot(self):
        def tg_handle_message_loop():
            self.updater = Updater(self.conf.get('api_token'))

            # Get the dispatcher to register handlers
            dp = self.updater.dispatcher

            # on different commands - answer in Telegram
            def get_chat_id(bot, update):
                self.logger.info("Received /getid from @{} ({} {}), chat id is {}.".format(
                    update.message.from_user.username,
                    update.message.from_user.first_name, update.message.from_user.last_name,
                    update.message.chat_id
                ))
                bot.send_message(
                    update.message.chat_id,
                    text='Your Chat ID is **{}**.'.format(
                        update.message.chat_id),
                    parse_mode=telegram.ParseMode.MARKDOWN
                )

            # callback for event actions
            def handle_event_actions(bot, update, **kwargs):
                query = update.callback_query
                chat_id = query.message.chat_id
                query_data = query.data.split(" ")

                self.logger.info("Received callback from @{} ({} {}).".format(
                    query.from_user.username,
                    query.from_user.first_name, query.from_user.last_name,
                ))

                if len(query_data) != 2:
                    query.answer(text='Invalid query.')
                    return

                action, event_id = query_data

                if action == 'mute' or action == 'acknowledge':
                    try:
                        new = self.state.alarms.transit(
                            id=event_id, action='TOGGLE_ACK')
                        query.answer(text='State has been changed to {}.'.format(
                            new
                        ))
                    except Exception as e:
                        query.answer(text='Unable to change state of event {}: {}'.format(
                            event_id, e
                        ))

                elif action == 'resolve':
                    self.state.alarms.transit(id=event_id, action='RESOLVE')
                    query.answer(text='Resolve.')

                elif action == 'details':
                    query.answer(text='Processing ...')
                    event = self.state.alarms.get_event(event_id)

                    if not event:
                        query.answer(text='Event not found.')

                    msg = "\n".join([
                        u'{icon} <b>[P{level}]</b> {title}',
                        u'<pre>{}</pre>',
                    ]).format(
                        "\n".join([
                            u"{}: {}".format(k.capitalize(), v)
                            for k, v in event.items()
                            if k != 'users'
                        ]),
                        icon=u'😱' if event['status'] in (
                            'PROBLEM', 'EVENT') else u'😅',
                        **event
                    )

                    self.updater.bot.send_message(
                        chat_id, text=msg,
                        parse_mode=telegram.ParseMode.HTML
                    )
                else:
                    query.answer(text='Invalid action.')

            dp.add_handler(CommandHandler("getid", get_chat_id))
            dp.add_handler(CallbackQueryHandler(
                handle_event_actions,
                pass_user_data=True, pass_chat_data=True))

            # log all errors
            def error(bot, update, error):
                self.logger.warn(
                    'Update "{}" caused error "{}"'.format(update, error))

            dp.add_error_handler(error)

            # Start the Bot
            self.updater.start_polling()

        tg_handle_message_loop()

    def shutdown(self):
        super(TelegramBackend, self).shutdown()
        self.updater.stop()

    @staticmethod
    def make_button(action, event):
        action = str(action)
        return telegram.InlineKeyboardButton(
            text=action.capitalize(),
            callback_data="{} {}".format(action.lower(), event['id'])
        )

    def send(self, ev):
        for user in ev['users']:
            chat_id = user.get('tg_chat_id', '')
            if not chat_id:
                continue

            buttons = telegram.InlineKeyboardMarkup([[
                TelegramBackend.make_button("acknowledge", ev),
                TelegramBackend.make_button("resolve", ev),
                TelegramBackend.make_button("details", ev),
            ]])

            msg = "\n".join([
                u'{icon} <b>[P{level}]</b> {title}',
                u'<pre>{text}</pre>',
            ]).format(
                icon=u'😱' if ev['status'] in (
                    'PROBLEM', 'EVENT') else u'😅',
                **ev
            )

            self.updater.bot.send_message(
                chat_id, text=msg,
                reply_markup=buttons,
                parse_mode=telegram.ParseMode.HTML
            )

EXPORT = TelegramBackend
