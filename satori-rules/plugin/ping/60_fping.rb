#!/usr/bin/env ruby

require 'net/http'
require 'json'
require 'open3'

# STDERR.puts Time.now.strftime('%F_%T')

KEY_ROOT = '/var/lib/puppet/ssl'
HOSTNAME = `hostname`.chomp()
metrics  = []

SSL_OPTIONS = {
    use_ssl: true,
    verify_mode: OpenSSL::SSL::VERIFY_NONE,
    keep_alive_timeout: 10,
    cert: OpenSSL::X509::Certificate.new( File.read( Dir.glob("#{KEY_ROOT}/certs/#{HOSTNAME}*.pem").first )),
    key:  OpenSSL::PKey::RSA.new( File.read( Dir.glob("#{KEY_ROOT}/private_keys/#{HOSTNAME}*.pem").first ))
}

http = Net::HTTP.start( 'puppet', 9081, SSL_OPTIONS)
response = http.request Net::HTTP::Get.new  '/v3/facts?query=%5B%22%3D%22%2C%20%22name%22%2C%20%22hostname%22%5D'
hosts = JSON.parse( response.body ).collect{ |host| host['value'] }
# STDERR.puts "ping hosts count: #{hosts.count}"

o, _ = Open3.capture2e("/usr/bin/timeout -k 3 --preserve-status 40s /usr/bin/fping  -c 10 -r 0 -i 10 -q -s -t 50", :stdin_data => hosts.join("\n"))

ts   = Time.now.strftime('%s').to_i
metric_template = {
    :timestamp => ts,
    :step      => 60,
}

# redis-push-stats    : xmt/rcv/%loss = 10/10/0%, min/avg/max = 0.40/0.50/0.72
FPING_SPLIT_REGEX = /^([a-z0-9\.\-_]+)[ ]+: xmt\/rcv\/%loss = ([0-9]+)\/([0-9]+)\/([0-9]+)%(, min\/avg\/max = ([0-9\.]+)\/([0-9\.]+)\/([0-9\.]+))?$/
o.split("\n").each do |line|
  if line.chomp =~ FPING_SPLIT_REGEX
    to = $1
    loss = $4.to_f
    latency = $6.to_f
    { 'ping.loss'    => loss,
      'ping.latency' => latency,
      'ping.alive'   => loss == 100.0 ? 0 : 1,
    }.map { |k,v|
      metrics.push( metric_template.merge({ :metric => k,
                                            :endpoint => to,
                                            :value => v,
                                            :tags => { :to => to,
                                                       :from => HOSTNAME }} ))
    }
  end
end

STDOUT.puts metrics.to_json
