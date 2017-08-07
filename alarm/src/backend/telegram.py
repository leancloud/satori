# -*- coding: utf-8 -*-
from __future__ import absolute_import

# -- stdlib --
# -- third party --
from telegram.ext import CommandHandler, Updater
import telegram

# -- own --
from backend.common import Backend, register_backend


# -- code --
@register_backend
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
                    text='Your Chat ID is **{}**.'.format(update.message.chat_id),
                    parse_mode=telegram.ParseMode.MARKDOWN
                )
            dp.add_handler(CommandHandler("getid", get_chat_id))

            # log all errors
            def error(bot, update, error):
                self.logger.warn('Update "{}" caused error "{}"'.format(update, error))
            dp.add_error_handler(error)

            # Start the Bot
            self.updater.start_polling()

        tg_handle_message_loop()

    def shutdown(self):
        super(TelegramBackend, self).shutdown()
        self.updater.stop()

    def send(self, user, event):
        chat_id = user.get('tg_chat_id', '')
        if not chat_id:
            return

        msg = u'{} **[P{}]**\n{}\n'.format(
            u'😱' if event['status'] in ('PROBLEM', 'EVENT') else u'😅',
            event['level'],
            event['title'],
        ) + event['text']

        self.updater.bot.send_message(
            chat_id, text=msg,
            parse_mode=telegram.ParseMode.MARKDOWN
        )
