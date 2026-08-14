module DnsTestHelper
  extend ActiveSupport::Concern

  WEB_PUSH_PUBLIC_TEST_IP = "142.250.185.206" # stable public IP for web push DNS stubs in tests
  THREAD_KEY = :dns_resolution_ips

  module ResolvDnsOpenStub
    def open(*, **)
      ips = Thread.current[DnsTestHelper::THREAD_KEY]
      if ips
        dns = Object.new
        dns.define_singleton_method(:each_address) do |_hostname, &block|
          ips.each { |ip| block.call(ip) }
        end
        yield dns
      else
        super
      end
    end
  end

  included do
    teardown { Thread.current[THREAD_KEY] = nil }
  end

  private
    def stub_dns_resolution(*ips)
      Thread.current[THREAD_KEY] = ips
    end

    def stub_web_push_dns_resolution
      stub_dns_resolution(WEB_PUSH_PUBLIC_TEST_IP)
    end
end

Resolv::DNS.singleton_class.prepend(DnsTestHelper::ResolvDnsOpenStub)
