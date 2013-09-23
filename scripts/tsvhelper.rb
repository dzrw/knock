# encoding: utf-8

h = {}

ARGF.each do |line|
  line.strip!.tr!("\t", '')

  m = /^\s*(.*):\s*(\d+(\.\d+)?)(μs)?$/.match(line)
  if m
    k = case m[1]
        when "Run Time (s)"; "Time"
        when "Throughput (ops/sec)"; "Tput"
        when "Mean Response Time (μs)"; next
        when "Load Efficiency (%)"; "Efcy"
        when "5th Percentile"; "0.05"
        when "95th Percentile"; "0.95"
        when "99th Percentile"; "0.99"
        else m[1]
        end

    h[k] = Float(m[2])
  end
end

puts h.keys.join(",")
puts h.values.join(",")
